package integration_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	qstsv1a1 "code.cloudfoundry.org/cf-operator/pkg/kube/apis/quarksstatefulset/v1alpha1"
	utils "code.cloudfoundry.org/quarks-utils/testing/integration"
	"code.cloudfoundry.org/quarks-utils/testing/machine"
	helper "code.cloudfoundry.org/quarks-utils/testing/testhelper"
)

var _ = Describe("QuarksStatefulSetActivePassive", func() {
	var (
		podNameByIndex = func(podName, index string) string {
			return fmt.Sprintf("%s-%s", podName, index)
		}
		passiveEvent = func(podName, index string) string {
			return fmt.Sprintf("pod %s promoted to passive", fmt.Sprintf("%s-%s", podName, index))
		}
		activeEvent = func(podName, index string) string {
			return fmt.Sprintf("pod %s promoted to active", fmt.Sprintf("%s-%s", podName, index))
		}

		probeValidation = func(podName, index string) string {
			return fmt.Sprintf("validating probe in active pod: %s", fmt.Sprintf("%s-%s", podName, index))
		}

		qStsName, podDesignationLabel, defaultPodLabel, eventReason, patchPath, patchValue, patchOp string
	)

	BeforeEach(func() {
		// Values required to define a patch mechanism
		patchPath = fmt.Sprintf("%s%s%s", "/metadata/labels/quarks.cloudfoundry.org", "~1", "pod-designation")
		patchValue = "true"
		patchOp = "add"
		eventReason = "active-passive"
		qStsName = fmt.Sprintf("test-ap-qsts-%s", helper.RandString(5))
		podDesignationLabel = "quarks.cloudfoundry.org/pod-designation=active"
		defaultPodLabel = "testpod=yes"
	})

	AfterEach(func() {
		Expect(env.WaitForPodsDelete(env.Namespace)).To(Succeed())
		// Skipping wait for PVCs to be deleted until the following is fixed
		// https://www.pivotaltracker.com/story/show/166896791
		// Expect(env.WaitForPVCsDelete(env.Namespace)).To(Succeed())
	})

	Context("when pod-designation label is not present", func() {
		sleepCMD := []string{"/bin/sh", "-c", "sleep 2"}
		It("should label a single pod out of one", func() {
			By("Creating a QuarksStatefulSet with a valid CRD probe cmd")
			var qSts *qstsv1a1.QuarksStatefulSet
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithProbeSinglePod(
				qStsName,
				sleepCMD,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Waiting for pod with index 0 to become active")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					qSts.ObjectMeta.Name,
					string(qSts.ObjectMeta.UID),
					eventReason,
					activeEvent(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait for pod with pod-designation label to be ready")
			err = env.WaitForPods(env.Namespace, podDesignationLabel)
			Expect(err).NotTo(HaveOccurred())

			By("Checking that only one pod is active and is the first one")
			pLB, err := env.GetPodNamesByLabel(env.Namespace, podDesignationLabel)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pLB)).To(Equal(1))
		})

		It("should not label if the probe fails", func() {
			By("Creating an QuarksStatefulSet")
			var qSts *qstsv1a1.QuarksStatefulSet
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithProbeMultiplePods(
				qStsName,
				[]string{"/bin/sh", "-c", "sleeps 2"},
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking for pods with default label")
			err = env.WaitForPods(env.Namespace, defaultPodLabel)
			Expect(err).NotTo(HaveOccurred())

			By("Checking that none pods are active")
			pLB, err := env.GetPodNamesByLabel(env.Namespace, podDesignationLabel)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pLB)).To(Equal(0))
		})

	})

	Context("when pod-designation label is present in one pod", func() {

		sleepCMD := []string{"/bin/sh", "-c", "sleep 2"}

		It("should remain with an active labelled pod", func() {
			By("Creating a QuarksStatefulSet with a valid CRD probe cmd")
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithActiveSinglePod(
				qStsName,
				sleepCMD,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking controller events for probe validation")
			objectName := qSts.ObjectMeta.Name
			objectUID := string(qSts.ObjectMeta.UID)
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					probeValidation(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait for active pod to be ready")
			err = env.WaitForPods(env.Namespace, podDesignationLabel)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when pod-designation label is present in one pod and probe fails", func() {

		cmdSleepTypo := []string{"/bin/sh", "-c", "sleeps 2"}

		It("should ensure all pods are pasive", func() {
			By("Creating a QuarksStatefulSet with pods that contain a wrong command")
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithProbeMultiplePods(
				qStsName,
				cmdSleepTypo,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Waiting for pod with index 0 to be ready")
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "0"))
			Expect(err).NotTo(HaveOccurred())

			By("Adding the pod-designation label to pod with index 0")
			err = env.PatchPod(env.Namespace, podNameByIndex(qStsName, "0"), patchOp, patchPath, patchValue)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for active/passive controller to reconcile when pod is labelled as active")
			objectName := qSts.ObjectMeta.Name
			objectUID := string(qSts.ObjectMeta.UID)
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					probeValidation(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for active/passive controller to remove the pod-designation label")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					passiveEvent(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking that no pods are marked as active")
			pLB, err := env.GetPodNamesByLabel(env.Namespace, podDesignationLabel)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pLB)).To(Equal(0))
		})
	})

	Context("when pod-designation label is present in multiple pods and only one probe pass", func() {
		// two set of cmds, one that runs as the CRD probe
		// the second one, runs as a patch, so that the next CRD probe executiong will pass
		cmdCatScript := []string{"/bin/sh", "-c", "cat /tmp/busybox-script.sh"}
		cmdTouchScript := []string{"/bin/sh", "-c", "touch /tmp/busybox-script.sh"}
		// pod container name
		containerName := "busybox"

		It("should ensure only one pod is active", func() {
			By("Creating a QuarksStatefulSet with pods that contain a probe that will initially fail")
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithProbeMultiplePods(
				qStsName,
				cmdCatScript,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Waiting for all pods owned by the qsts to be ready")
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "0"))
			Expect(err).NotTo(HaveOccurred())
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "1"))
			Expect(err).NotTo(HaveOccurred())
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "2"))
			Expect(err).NotTo(HaveOccurred())

			By("Adding the pod-designation label to all pods")
			err = env.PatchPod(env.Namespace, podNameByIndex(qStsName, "0"), patchOp, patchPath, patchValue)
			Expect(err).NotTo(HaveOccurred())
			err = env.PatchPod(env.Namespace, podNameByIndex(qStsName, "1"), patchOp, patchPath, patchValue)
			Expect(err).NotTo(HaveOccurred())
			err = env.PatchPod(env.Namespace, podNameByIndex(qStsName, "2"), patchOp, patchPath, patchValue)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for active/passive controller to remove the labels when the probe fails")
			objectName := qSts.ObjectMeta.Name
			objectUID := string(qSts.ObjectMeta.UID)
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					passiveEvent(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					passiveEvent(qStsName, "1"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking controller events")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					passiveEvent(qStsName, "2"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Executing in pod with index 1 a cmd to force the probe to pass")
			kubeConfig, err := utils.KubeConfig()
			Expect(err).NotTo(HaveOccurred())
			kclient, err := kubernetes.NewForConfig(kubeConfig)
			Expect(err).NotTo(HaveOccurred())

			p, err := env.GetPod(env.Namespace, fmt.Sprintf("%s-1", qStsName))
			Expect(err).NotTo(HaveOccurred())

			ec, err := env.ExecPodCMD(
				kclient,
				kubeConfig,
				p,
				containerName,
				cmdTouchScript,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(ec).To(Equal(true))

			By("Waiting for pod with index 1 to become active")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					qSts.ObjectMeta.Name,
					string(qSts.ObjectMeta.UID),
					eventReason,
					activeEvent(qStsName, "1"),
				)
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when active passive pod fails a new one becomes active", func() {
		cmdCatScript := []string{"/bin/sh", "-c", "cat /tmp/busybox-script.sh"}
		cmdTouchScript := []string{"/bin/sh", "-c", "touch /tmp/busybox-script.sh"}
		containerName := "busybox"
		It("Creating a QuarksStatefulSet", func() {
			By("Defining a probe cmd that will fail")
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithProbeMultiplePods(
				qStsName,
				cmdCatScript,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Waiting for all pods owned by the qsts to be ready")
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "0"))
			Expect(err).NotTo(HaveOccurred())
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "1"))
			Expect(err).NotTo(HaveOccurred())
			err = env.WaitForPodReady(env.Namespace, podNameByIndex(qStsName, "2"))
			Expect(err).NotTo(HaveOccurred())

			By("Executing a cmd in pod index 0 to make the probe successful")
			kubeConfig, err := utils.KubeConfig()
			Expect(err).NotTo(HaveOccurred())
			kclient, err := kubernetes.NewForConfig(kubeConfig)
			Expect(err).NotTo(HaveOccurred())

			p, err := env.GetPod(env.Namespace, fmt.Sprintf("%s-0", qStsName))
			Expect(err).NotTo(HaveOccurred())
			ec, err := env.ExecPodCMD(
				kclient,
				kubeConfig,
				p,
				containerName,
				cmdTouchScript,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(ec).To(Equal(true))

			By("Waiting for an event on a new active pod with index 0")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					qSts.ObjectMeta.Name,
					string(qSts.ObjectMeta.UID),
					eventReason,
					activeEvent(qStsName, "0"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

			By("Excuting a cmd to make the active pod fail its probe")
			p, err = env.GetPod(env.Namespace, fmt.Sprintf("%s-0", qStsName))
			Expect(err).NotTo(HaveOccurred())
			ec, err = env.ExecPodCMD(
				kclient,
				kubeConfig,
				p,
				containerName,
				[]string{"/bin/sh", "-c", "rm /tmp/busybox-script.sh"},
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(ec).To(Equal(true))

			By("Executing a cmd in another pod to make the probe successful")
			p, err = env.GetPod(env.Namespace, fmt.Sprintf("%s-1", qStsName))
			Expect(err).NotTo(HaveOccurred())
			ec, err = env.ExecPodCMD(
				kclient,
				kubeConfig,
				p,
				containerName,
				cmdTouchScript,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(ec).To(Equal(true))

			By("Looking for an event on a new active pod")
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					qSts.ObjectMeta.Name,
					string(qSts.ObjectMeta.UID),
					eventReason,
					activeEvent(qStsName, "1"),
				)
			})
			Expect(err).NotTo(HaveOccurred())

		})
	})

	Context("when CRD does not specify a probe periodSeconds", func() {
		cmdDate := []string{"/bin/sh", "-c", "date"}
		It("should ensure the proper event takes place", func() {
			By("Creating a QuarksStatefulSet with pods that contain a probe that will initially fail")
			qSts, tearDown, err := env.CreateQuarksStatefulSet(env.Namespace, env.QstsWithoutProbeMultiplePods(
				qStsName,
				cmdDate,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(qSts).NotTo(Equal(nil))
			defer func(tdf machine.TearDownFunc) { Expect(tdf()).To(Succeed()) }(tearDown)

			By("Checking events to match the default periodSeconds")
			objectName := qSts.ObjectMeta.Name
			objectUID := string(qSts.ObjectMeta.UID)
			err = wait.PollImmediate(5*time.Second, 35*time.Second, func() (bool, error) {
				return env.GetNamespaceEvents(env.Namespace,
					objectName,
					objectUID,
					eventReason,
					"periodSeconds probe was not specified, going to default to 10 secs",
				)
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
