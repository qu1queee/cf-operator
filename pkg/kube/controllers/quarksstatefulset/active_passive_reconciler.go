package quarksstatefulset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	qstsv1a1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/quarksstatefulset/v1alpha1"
	"code.cloudfoundry.org/quarks-utils/pkg/config"
	"code.cloudfoundry.org/quarks-utils/pkg/ctxlog"
	podutil "code.cloudfoundry.org/quarks-utils/pkg/pod"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	crc "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NewActivePassiveReconciler returns a new reconcile.Reconciler for active/passive controller
func NewActivePassiveReconciler(ctx context.Context, config *config.Config, mgr manager.Manager, kclient kubernetes.Interface) reconcile.Reconciler {
	return &ReconcileStatefulSetActivePassive{
		ctx:        ctx,
		config:     config,
		client:     mgr.GetClient(),
		kclient:    kclient,
		scheme:     mgr.GetScheme(),
		kubeconfig: mgr.GetConfig(),
	}
}

// ReconcileStatefulSetActivePassive reconciles an QuarksStatefulSet object when references changes
type ReconcileStatefulSetActivePassive struct {
	ctx        context.Context
	client     crc.Client
	kclient    kubernetes.Interface
	scheme     *runtime.Scheme
	config     *config.Config
	kubeconfig *restclient.Config
}

// Reconcile reads that state of the cluster for a QuarksStatefulSet object
// and makes changes based on the state read and what is in the QuarksStatefulSet.Spec
// Note:
// The Reconcile Loop will always requeue the request under completition. For this specific
// loop, the requeue will happen after the ActivePassiveProbe PeriodSeconds is reached.
func (r *ReconcileStatefulSetActivePassive) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	qSts := &qstsv1a1.QuarksStatefulSet{}

	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	ctxlog.Info(ctx, "Reconciling for active/passive QuarksStatefulSet", request.NamespacedName)

	if err := r.client.Get(ctx, request.NamespacedName, qSts); err != nil {
		if apierrors.IsNotFound(err) {
			// Reconcile successful - don´t requeue
			ctxlog.Infof(ctx, "Failed to find quarks statefulset '%s', not retrying: %s", request.NamespacedName, err)
			return reconcile.Result{}, nil
		}
		// Reconcile failed due to error - requeue
		ctxlog.Errorf(ctx, "Failed to get quarks statefulset '%s': %s", request.NamespacedName, err)
		return reconcile.Result{}, err
	}

	statefulSetVersions, err := r.listStatefulSetVersions(ctx, qSts)
	if err != nil {
		return reconcile.Result{}, err
	}
	maxAvailableVersion := qSts.GetMaxAvailableVersion(statefulSetVersions)

	statefulSets, err := listStatefulSetsFromInformer(ctx, r.client, qSts)
	if err != nil {
		// Reconcile failed due to error - requeue
		return reconcile.Result{}, errors.Wrapf(err, "couldn't list StatefulSets for active/passive reconciliation.")
	}

	// Retrieve the last versioned statefulset, based on the max available version
	desiredSts, err := getDesiredSts(statefulSets, fmt.Sprintf("%d", maxAvailableVersion))
	if err != nil {
		// Reconcile failed due to error - requeue
		return reconcile.Result{}, errors.Wrapf(err, "couldn´t get the last versioned statefulset owned by QuarksStatefulSet")
	}

	// ctxlog.Info(ctx, "Max available version for quarks sts'", qSts.Name, "' is version '", maxAvailableVersion, "' in namespace '", qSts.Namespace, "'.")
	ownedPods, err := r.getStsPodList(ctx, desiredSts)
	if err != nil {
		// Reconcile failed due to error - requeue
		return reconcile.Result{}, errors.Wrapf(err, "couldn't retrieve pod items from sts: %s", desiredSts.Name)
	}

	// retrieves the ActivePassiveProbe children key,
	// this is the container name in where the ActivePassiveProbe
	// cmd, needs to be executed
	containerName, err := getProbeContainerName(qSts.Spec.ActivePassiveProbe)
	if err != nil {
		// Reconcile failed due to error - requeue
		return reconcile.Result{}, errors.Wrapf(err, "None container name found in probe for %s QuarksStatefulSet", qSts.Name)
	}

	err = r.defineActiveContainer(ctx, containerName, ownedPods, qSts)
	if err != nil {
		// Reconcile failed due to error - requeue
		return reconcile.Result{}, err
	}

	periodSeconds := time.Second * time.Duration(qSts.Spec.ActivePassiveProbe[containerName].PeriodSeconds)
	if periodSeconds == (time.Second * time.Duration(0)) {
		ctxlog.WithEvent(qSts, "active-passive").Debugf(ctx, "periodSeconds probe was not specified, going to default to 10 secs")
		periodSeconds = time.Second * 5
	}

	// Reconcile for any reason than error after the ActivePassiveProbe PeriodSeconds
	return reconcile.Result{RequeueAfter: periodSeconds}, nil
}

func (r *ReconcileStatefulSetActivePassive) defineActiveContainer(ctx context.Context, container string, pods *corev1.PodList, qSts *qstsv1a1.QuarksStatefulSet) (err error) {
	var p *corev1.Pod
	labelledPods := r.getActiveLabels(pods, container)

	probeCmd := qSts.Spec.ActivePassiveProbe[container].Exec.Command

	// switch depending on pods with active label
	switch pl := len(labelledPods); {
	// Nothing have been labeled yet
	case pl == 0:
		ctxlog.WithEvent(qSts, "active-passive").Debugf(
			ctx,
			"there is no active pod",
		)

		p, err = r.getActivePod(pods, container, probeCmd)
		if err != nil {
			return err
		}

		podLabels := p.GetLabels()
		if podLabels == nil {
			podLabels = map[string]string{}

		}

		podLabels[qstsv1a1.LabelActiveContainer] = "active"

		err = r.client.Update(r.ctx, p)
		if err != nil {
			return err
		}
		ctxlog.WithEvent(qSts, "active-passive").Debugf(
			ctx,
			"pod %s promoted to active",
			p.Name,
		)
	case pl > 0:
		for _, p := range labelledPods {
			ctxlog.WithEvent(qSts, "active-passive").Debugf(
				ctx,
				"validating probe in active pod: %s",
				p.Name,
			)
			if probeSucceeded, _ := r.execContainerCmd(&p, container, probeCmd); !probeSucceeded {
				l := p.GetLabels()
				delete(l, qstsv1a1.LabelActiveContainer)
				err = r.client.Update(r.ctx, &p)
				if err != nil {
					return err
				}
				ctxlog.WithEvent(qSts, "active-passive").Debugf(
					ctx,
					"pod %s promoted to passive",
					p.Name,
				)
			}
		}
	}
	return nil
}

func (r *ReconcileStatefulSetActivePassive) getActivePod(pods *corev1.PodList, container string, cmd []string) (*corev1.Pod, error) {
	for _, pod := range pods.Items {
		if succeed, _ := r.execContainerCmd(&pod, container, cmd); succeed && podutil.IsPodReady(&pod) {
			return &pod, nil
		}
	}
	return nil, errors.Errorf("executing cmd: %v, in container: %s, failed\n", cmd, container)
}

func (r *ReconcileStatefulSetActivePassive) getActiveLabels(pods *corev1.PodList, container string) []corev1.Pod {
	var labelledPods []corev1.Pod
	for _, pod := range pods.Items {
		labels := pod.GetLabels()
		if _, found := labels[qstsv1a1.LabelActiveContainer]; found {
			labelledPods = append(labelledPods, pod)
		}
	}
	return labelledPods
}

func (r *ReconcileStatefulSetActivePassive) execContainerCmd(pod *corev1.Pod, container string, command []string) (bool, error) {
	req := r.kclient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    false,
		}, clientscheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(r.kubeconfig, "POST", req.URL())
	if err != nil {
		return false, errors.New("failed to initialize remote command executor")
	}
	if err = executor.Stream(remotecommand.StreamOptions{Stdin: os.Stdin, Stdout: os.Stdout, Tty: false}); err != nil {
		return false, errors.Wrapf(err, "failed executing command in pod: %s, container: %s in namespace: %s",
			pod.Name,
			container,
			pod.Namespace,
		)
	}
	ctxlog.Info(r.ctx, "Succesfully exec cmd in container: ", container, ", inside pod: ", pod.Name)

	return true, nil
}

func (r *ReconcileStatefulSetActivePassive) getStsPodList(ctx context.Context, desiredSts *appsv1.StatefulSet) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	err := r.client.List(ctx,
		podList,
		crc.InNamespace(desiredSts.Namespace),
	)
	if err != nil {
		return nil, err
	}
	return podList, nil
}

func (r *ReconcileStatefulSetActivePassive) listStatefulSetVersions(ctx context.Context, qStatefulSet *qstsv1a1.QuarksStatefulSet) (map[int]bool, error) {
	ctxlog.Debug(ctx, "Listing StatefulSets owned by QuarksStatefulSet '", qStatefulSet.Name, "'.")

	versions := map[int]bool{}

	statefulSets, err := listStatefulSetsFromInformer(ctx, r.client, qStatefulSet)
	if err != nil {
		return nil, err
	}

	for _, statefulSet := range statefulSets {
		strVersion, found := statefulSet.Annotations[qstsv1a1.AnnotationVersion]
		if !found {
			return versions, errors.Errorf("version annotation is not found from: %+v", statefulSet.Annotations)
		}

		version, err := strconv.Atoi(strVersion)
		if err != nil {
			return versions, errors.Wrapf(err, "version annotation is not an int: %s", strVersion)
		}

		ready, err := r.isStatefulSetReady(ctx, &statefulSet)
		if err != nil {
			return nil, err
		}

		versions[version] = ready
	}

	return versions, nil
}

func (r *ReconcileStatefulSetActivePassive) isStatefulSetReady(ctx context.Context, statefulSet *appsv1.StatefulSet) (bool, error) {
	podLabels := map[string]string{appsv1.StatefulSetRevisionLabel: statefulSet.Status.CurrentRevision}

	podList := &corev1.PodList{}
	err := r.client.List(ctx,
		podList,
		crc.InNamespace(statefulSet.Namespace),
		crc.MatchingLabels(podLabels),
	)

	if err != nil {
		return false, err
	}

	for _, pod := range podList.Items {
		if metav1.IsControlledBy(&pod, statefulSet) {
			if podutil.IsPodReady(&pod) {
				ctxlog.Debugf(ctx, "Pod '%s' owned by StatefulSet '%s' is running.", pod.Name, statefulSet.Name)
				return true, nil
			}
		}
	}

	return false, nil
}

func getDesiredSts(sts []appsv1.StatefulSet, maxVersion string) (*appsv1.StatefulSet, error) {
	dS := &appsv1.StatefulSet{}
	for _, statefulSet := range sts {
		strVersion, found := statefulSet.Annotations[qstsv1a1.AnnotationVersion]
		if !found {
			return nil, errors.New("non versioned statefulset found")
		}
		if strVersion == string(maxVersion) {
			if statefulSet.Spec.Replicas != nil {
				return &statefulSet, nil
			}
		}

	}
	return dS, errors.New("non desired statefulset found")
}

func getProbeContainerName(p map[string]*corev1.Probe) (string, error) {
	for key := range p {
		return key, nil
	}
	return "", errors.New("failed to find a container key in the active/passive probe in the current QuarksStatefulSet")
}

// KubeConfig returns a kube config for this environment
func KubeConfig() (*rest.Config, error) {
	location := os.Getenv("KUBECONFIG")
	if location == "" {
		location = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", location)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
