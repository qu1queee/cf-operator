package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deploy", func() {
	Context("when correctly setup", func() {
		It("should deploy a pod", func() {
			tearDown, err := env.CreateConfigMap(env.Namespace, env.DefaultBOSHManifest("manifest"))
			Expect(err).NotTo(HaveOccurred())
			defer tearDown()

			tearDown, err = env.CreateFissileCR(env.Namespace, env.DefaultFissileCR("test", "manifest"))
			Expect(err).NotTo(HaveOccurred())
			defer tearDown()

			// check for pod
			err = env.WaitForPod("diego-pod")
			Expect(err).NotTo(HaveOccurred(), "error waiting for pod from deployment")
		})
	})

})
