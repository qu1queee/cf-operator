package machinery

import (
	fisv1 "code.cloudfoundry.org/cf-operator/pkg/apis/fissile/v1alpha1"
	"code.cloudfoundry.org/cf-operator/pkg/client/clientset/versioned"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Machine produces and destroys resources for tests
type Machine struct {
	Clientset          *kubernetes.Clientset
	VersionedClientset *versioned.Clientset
}

// TearDownFunc tears down the resource
type TearDownFunc func()

// CreateConfigMap creates a ConfigMap and returns a function to delete it
func (m *Machine) CreateConfigMap(namespace string, configMap corev1.ConfigMap) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().ConfigMaps(namespace)
	_, err := client.Create(&configMap)
	return func() {
		client.Delete(configMap.GetName(), &v1.DeleteOptions{})
	}, err
}

// CreateSecret creates a secret and returns a function to delete it
func (m *Machine) CreateSecret(namespace string, secret corev1.Secret) (TearDownFunc, error) {
	client := m.Clientset.CoreV1().Secrets(namespace)
	_, err := client.Create(&secret)
	return func() {
		client.Delete(secret.GetName(), &v1.DeleteOptions{})
	}, err
}

// CreateFissileCR creates a BOSHDeployment custom resource and returns a function to delete it
func (m *Machine) CreateFissileCR(namespace string, deployment fisv1.BOSHDeployment) (*fisv1.BOSHDeployment, TearDownFunc, error) {
	client := m.VersionedClientset.Fissile().BOSHDeployments(namespace)
	d, err := client.Create(&deployment)
	return d, func() {
		client.Delete(deployment.GetName(), &v1.DeleteOptions{})
	}, err
}

// UpdateFissileCR creates a BOSHDeployment custom resource and returns a function to delete it
func (m *Machine) UpdateFissileCR(namespace string, deployment fisv1.BOSHDeployment) (*fisv1.BOSHDeployment, TearDownFunc, error) {
	client := m.VersionedClientset.Fissile().BOSHDeployments(namespace)
	d, err := client.Update(&deployment)
	return d, func() {
		client.Delete(deployment.GetName(), &v1.DeleteOptions{})
	}, err
}

// DefaultBOSHManifest for tests
func (m *Machine) DefaultBOSHManifest(name string) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{Name: name},
		Data: map[string]string{
			"manifest": `instance-groups:
- name: diego
  instances: 3
- name: myqsl
`,
		},
	}
}

// DefaultSecret for tests
func (m *Machine) DefaultSecret(name string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: v1.ObjectMeta{Name: name},
		StringData: map[string]string{},
	}
}

// DefaultFissileCR fissile deployment CR
func (m *Machine) DefaultFissileCR(name, manifestRef string) fisv1.BOSHDeployment {
	return fisv1.BOSHDeployment{
		ObjectMeta: v1.ObjectMeta{Name: name},
		Spec: fisv1.BOSHDeploymentSpec{
			ManifestRef: manifestRef,
		},
	}
}
