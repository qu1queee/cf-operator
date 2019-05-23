package boshdeployment

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/bpm"
	bdm "code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
	bdv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	ejv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedjob/v1alpha1"
	estsv1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/extendedstatefulset/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/config"
	log "code.cloudfoundry.org/cf-operator/pkg/kube/util/ctxlog"
)

var _ reconcile.Reconciler = &ReconcileBOSHDeployment{}

// NewBPMReconciler returns a new reconcile.Reconciler
func NewBPMReconciler(ctx context.Context, config *config.Config, mgr manager.Manager, resolver bdm.Resolver, srf setReferenceFunc, kubeConverter *bdm.KubeConverter) reconcile.Reconciler {
	return &ReconcileBPM{
		ctx:           ctx,
		config:        config,
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		resolver:      resolver,
		setReference:  srf,
		kubeConverter: kubeConverter,
	}
}

// ReconcileBPM reconciles an Instance Group BPM versioned secret
type ReconcileBPM struct {
	ctx           context.Context
	config        *config.Config
	client        client.Client
	scheme        *runtime.Scheme
	resolver      bdm.Resolver
	setReference  setReferenceFunc
	kubeConverter *bdm.KubeConverter
}

// Reconcile reconciles an Instance Group BPM versioned secret read the corresponding
// desired manifest. It then applies BPM information and deploys instance groups.
func (r *ReconcileBPM) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Set the ctx to be Background, as the top-level context for incoming requests.
	ctx, cancel := context.WithTimeout(r.ctx, r.config.CtxTimeOut)
	defer cancel()

	log.Infof(ctx, "Reconciling Instance Group BPM versioned secret '%s'", request.NamespacedName)
	bpmSecret := &corev1.Secret{}
	err := r.client.Get(ctx, request.NamespacedName, bpmSecret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Debug(ctx, "Skip reconcile: Instance Group BPM versioned secret not found")
			return reconcile.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return reconcile.Result{
				Requeue:      true,
				RequeueAfter: time.Second * 5,
			},
			log.WithEvent(bpmSecret, "GetBPMSecret").Errorf(ctx, "Failed to get Instance Group BPM versioned secret '%s': %v", request.NamespacedName, err)
	}

	// Get the label from the BPM Secret and read the corresponding desired manifest
	var boshDeploymentName string
	var ok bool
	if boshDeploymentName, ok = bpmSecret.Labels[bdv1.LabelDeploymentName]; !ok {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "GetBOSHDeploymentLabel").Errorf(ctx, "There's no label for a BOSH Deployment name on the Instance Group BPM versioned bpmSecret '%s'", request.NamespacedName)
	}
	manifest, err := r.resolver.ReadDesiredManifest(ctx, boshDeploymentName, request.Namespace)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "DesiredManifestReadError").Errorf(ctx, "Failed to read desired manifest '%s': %v", request.NamespacedName, err)
	}

	// Apply BPM information
	instanceGroupName, ok := bpmSecret.Labels[ejv1.LabelPersistentSecretContainer]
	if !ok {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "LabelMissingError").Errorf(ctx, "Missing container label for bpm information bpmSecret '%s'", request.NamespacedName)
	}

	resources, err := r.applyBPMResources(bpmSecret, manifest)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "BPMApplyingError").Errorf(ctx, "Failed to apply BPM information: %v", err)
	}

	// Start the instance groups referenced by this BPM secret
	err = r.deployInstanceGroups(ctx, bpmSecret, instanceGroupName, resources)
	if err != nil {
		return reconcile.Result{},
			log.WithEvent(bpmSecret, "InstanceGroupStartError").Errorf(ctx, "Failed to start : %v", err)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBPM) applyBPMResources(bpmSecret *corev1.Secret, manifest *bdm.Manifest) (*bdm.BPMResources, error) {
	resources := &bdm.BPMResources{}
	bpmInfo := map[string]bpm.Configs{}
	bpmConfigs := bpm.Configs{}

	instanceGroupName, ok := bpmSecret.Labels[ejv1.LabelPersistentSecretContainer]
	if !ok {
		return resources, errors.Errorf("Missing container label for bpm information secret '%s'", bpmSecret.Name)
	}

	err := yaml.Unmarshal(bpmSecret.Data["bpm.yaml"], &bpmConfigs)
	if err != nil {
		return resources, err
	}

	bpmInfo[instanceGroupName] = bpmConfigs
	resources, err = r.kubeConverter.BPMResources(manifest.Name, manifest.InstanceGroups, manifest, bpmInfo)
	if err != nil {
		return resources, err
	}

	return resources, nil
}

// deployInstanceGroups create or update ExtendedJobs and ExtendedStatefulSets for instance groups
func (r *ReconcileBPM) deployInstanceGroups(ctx context.Context, secret *corev1.Secret, instanceGroupName string, resources *bdm.BPMResources) error {
	log.Debugf(ctx, "Creating extendedJobs and extendedStatefulSets for secret group '%s'", instanceGroupName)

	for _, eJob := range resources.Errands {
		if eJob.Labels[bdm.LabelInstanceGroupName] != instanceGroupName {
			continue
		}

		if err := r.setReference(secret, &eJob, r.scheme); err != nil {
			return log.WithEvent(secret, "ExtendedJobForDeploymentError").Errorf(ctx, "Failed to set reference for ExtendedJob secret group '%s' : %v", instanceGroupName, err)
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.client, eJob.DeepCopy(), func(obj runtime.Object) error {
			if existingEJob, ok := obj.(*ejv1.ExtendedJob); ok {
				eJob.ObjectMeta.ResourceVersion = existingEJob.ObjectMeta.ResourceVersion
				eJob.DeepCopyInto(existingEJob)
				return nil
			}
			return fmt.Errorf("object is not an ExtendedJob")
		})
		if err != nil {
			return log.WithEvent(secret, "ApplyExtendedJobError").Errorf(ctx, "Failed to apply ExtendedJob for secret group '%s' : %v", instanceGroupName, err)
		}
	}

	for _, svc := range resources.Services {
		if svc.Labels[bdm.LabelInstanceGroupName] != instanceGroupName {
			continue
		}

		if err := r.setReference(secret, &svc, r.scheme); err != nil {
			return log.WithEvent(secret, "ServiceForDeploymentError").Errorf(ctx, "Failed to set reference for Service secret group '%s' : %v", instanceGroupName, err)
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.client, svc.DeepCopy(), func(obj runtime.Object) error {
			if existingSvc, ok := obj.(*corev1.Service); ok {
				// Should keep current ClusterIP and ResourceVersion when update
				svc.Spec.ClusterIP = existingSvc.Spec.ClusterIP
				svc.ObjectMeta.ResourceVersion = existingSvc.ObjectMeta.ResourceVersion
				svc.DeepCopyInto(existingSvc)
				return nil
			}
			return fmt.Errorf("object is not a Service")
		})
		if err != nil {
			return log.WithEvent(secret, "ApplyServiceError").Errorf(ctx, "Failed to apply Service for instance group '%s' : %v", instanceGroupName, err)
		}
	}

	for _, eSts := range resources.InstanceGroups {
		if eSts.Labels[bdm.LabelInstanceGroupName] != instanceGroupName {
			continue
		}

		if err := r.setReference(secret, &eSts, r.scheme); err != nil {
			return log.WithEvent(secret, "ExtendedStatefulSetForDeploymentError").Errorf(ctx, "Failed to set reference for ExtendedStatefulSet secret group '%s' : %v", instanceGroupName, err)
		}

		_, err := controllerutil.CreateOrUpdate(ctx, r.client, eSts.DeepCopy(), func(obj runtime.Object) error {
			if existingSts, ok := obj.(*estsv1.ExtendedStatefulSet); ok {
				eSts.ObjectMeta.ResourceVersion = existingSts.ObjectMeta.ResourceVersion
				eSts.DeepCopyInto(existingSts)
				return nil
			}
			return fmt.Errorf("object is not an ExtendStatefulSet")
		})
		if err != nil {
			return log.WithEvent(secret, "ApplyExtendedStatefulSetError").Errorf(ctx, "Failed to apply ExtendedStatefulSet for secret group '%s' : %v", instanceGroupName, err)
		}
	}

	return nil
}
