package integration_test

import (
	"fmt"

	fisv1 "code.cloudfoundry.org/cf-operator/pkg/apis/fissile/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lifecycle", func() {
	var (
		fissileCR   fisv1.BOSHDeployment
		newManifest corev1.ConfigMap
	)

	Context("when correctly setup", func() {
		BeforeEach(func() {
			fissileCR = env.DefaultFissileCR("test", "manifest")
			newManifest = corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{Name: "newmanifest"},
				Data: map[string]string{
					"manifest": `instance-groups:
- name: updated
  instances: 1
`,
				},
			}
		})

		It("should exercise a deployment lifecycle", func() {
			tearDown, err := env.CreateConfigMap(env.Namespace, env.DefaultBOSHManifest("manifest"))
			Expect(err).NotTo(HaveOccurred())
			defer tearDown()

			var versionedCR *fisv1.BOSHDeployment
			versionedCR, tearDown, err = env.CreateFissileCR(env.Namespace, fissileCR)
			Expect(err).NotTo(HaveOccurred())
			defer tearDown()

			err = env.WaitForPod("diego-pod")
			Expect(err).NotTo(HaveOccurred(), "error waiting for pod from initial deployment")

			// Update
			tearDown, err = env.CreateConfigMap(env.Namespace, newManifest)
			Expect(err).NotTo(HaveOccurred())
			defer tearDown()

			versionedCR.Spec.ManifestRef = "newmanifest"
			_, _, err = env.UpdateFissileCR(env.Namespace, *versionedCR)
			Expect(err).NotTo(HaveOccurred())

			err = env.WaitForPod("updated-pod")
			fmt.Printf("%+v", env.AllLogMessages())
			Expect(err).NotTo(HaveOccurred(), "error waiting for pod from updated deployment")

			// TODO after update we still have diego-pod around
			Expect(env.PodRunning("diego-pod")).To(BeTrue())
			Expect(env.PodRunning("updated-pod")).To(BeTrue())

			// TODO delete: use finalizers instead of Reconcile()?

			// delete
			// wait
			// check

		})
	})

})
