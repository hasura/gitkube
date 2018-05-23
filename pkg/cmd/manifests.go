package cmd

import (
	api "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var crd = apiextensionsv1beta1.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name: "remotes.gitkube.sh",
	},
	Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
		Group:   api.SchemeGroupVersion.Group,
		Version: api.SchemeGroupVersion.Version,
		Scope:   apiextensionsv1beta1.NamespaceScoped,
		Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
			Plural:     "remotes",
			Singular:   "remote",
			Kind:       "Remote",
			ShortNames: []string{"rem"},
		},
	},
}

var sa = corev1.ServiceAccount{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "gitkube",
		Namespace: "kube-system",
	},
}

var crb = rbacv1.ClusterRoleBinding{
	ObjectMeta: metav1.ObjectMeta{
		Name: "gitkube",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "gitkube",
			Namespace: "kube-system",
		},
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
	},
}

var cm = corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "gitkube-ci-conf",
		Namespace: "kube-system",
	},
}

var gitkubed = extensionsv1beta1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "gitkubed",
		Namespace: "kube-system",
	},
	Spec: extensionsv1beta1.DeploymentSpec{
		Replicas: int2ptr(1),
		Strategy: extensionsv1beta1.DeploymentStrategy{
			Type: extensionsv1beta1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &extensionsv1beta1.RollingUpdateDeployment{
				MaxUnavailable: &intstr.IntOrString{
					IntVal: 1,
				},
				MaxSurge: &intstr.IntOrString{
					IntVal: 1,
				},
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "gitkubed"},
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: "gitkubed",
				Containers: []corev1.Container{
					{
						Name:  "sshd",
						Image: "hasura/gitkubed:v0.1.1",
						Command: []string{
							"bash",
							"/sshd-lib/start_sshd.sh",
						},
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								Name:          "ssh",
								ContainerPort: 22,
								Protocol:      corev1.ProtocolTCP,
							},
						},
						ReadinessProbe: &corev1.Probe{
							InitialDelaySeconds: 5,
							PeriodSeconds:       2,
							// TCPSocket: &corev1.TCPSocketAction{
							// 	Port: intstr.IntOrString{
							// 		IntVal: 22,
							// 	},
							// },
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: "200m"},
								corev1.ResourceMemory: resource.Quantity{Format: "500Mi"},
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: "100m"},
								corev1.ResourceMemory: resource.Quantity{Format: "500Mi"},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "docker-sock",
								MountPath: "/var/run/docker.sock",
							},
							{
								Name:      "host-group",
								MountPath: "/hasura-data/group",
								ReadOnly:  true,
							},
							{
								Name:      "gitkube-ci-conf",
								MountPath: "/sshd-conf",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
				// {
				// 	Name: "docker-sock",
				// 	HostPath: &corev1.HostPathVolumeSource{
				// 		Path: "/var/run/docker.sock",
				// 	},
				// },
				// {
				// 	Name: "host-group",
				// 	HostPath: &corev1.HostPathVolumeSource{
				// 		Path: "/etc/group",
				// 	},
				// },
				// {
				// 	Name: "gitkube-ci-conf",
				// 	ConfigMap: &corev1.ConfigMapVolumeSource{
				// 		Name:        "gitkube-ci-conf",
				// 		DefaultMode: int2ptr(420),
				// 	},
				// },
				},
			},
		},
	},
}

func int2ptr(i int32) *int32 {
	return &i
}

func createCRD() error {
	clientset, err := apiextensionsclient.NewForConfig(kconfig)
	if err != nil {
		return errors.Wrap(err, "error creating apiextensionsclient")
	}
	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd)
	if err != nil {
		return errors.Wrap(err, "error creating crd")
	}
	return nil
}
