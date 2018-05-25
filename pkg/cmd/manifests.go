package cmd

import (
	api "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const CRDName = "remotes.gitkube.sh"

func newCRD() apiextensionsv1beta1.CustomResourceDefinition {
	return apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: CRDName,
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
}

const SAName = "gitkube"

func newSA(namespace string) corev1.ServiceAccount {
	return corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SAName,
			Namespace: namespace,
		},
	}

}

const CRBName = "gitkube"

func newCRB(namespace string) rbacv1.ClusterRoleBinding {
	return rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: CRBName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      SAName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
	}
}

const CMName = "gitkube-ci-conf"

func newCM(namespace string) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CMName,
			Namespace: namespace,
		},
	}

}

const GitkubedDeploymentName = "gitkubed"

func newGitkubed(namespace string) extensionsv1beta1.Deployment {
	return extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GitkubedDeploymentName,
			Namespace: namespace,
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
					ServiceAccountName: SAName,
					Containers: []corev1.Container{
						{
							Name:  "sshd",
							Image: "hasura/gitkubed:" + GetVersion(),
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
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											IntVal: 22,
										},
									},
								},
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
									Name:      CMName,
									MountPath: "/sshd-conf",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "docker-sock",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/run/docker.sock",
								},
							},
						},
						{
							Name: "host-group",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/etc/group",
								},
							},
						},
						{
							Name: CMName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: CMName,
									},
									DefaultMode: int2ptr(420),
								},
							},
						},
					},
				},
			},
		},
	}
}

const GitkubeControllerDeploymentName = "gitkube-controller"

func newGitkubeController(namespace string) extensionsv1beta1.Deployment {
	return extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GitkubeControllerDeploymentName,
			Namespace: namespace,
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
					Labels: map[string]string{"app": "gitkube-controller"},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: SAName,
					Containers: []corev1.Container{
						{
							Name:            "sshd",
							Image:           "hasura/gitkube-controller:" + GetVersion(),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.Quantity{Format: "50m"},
									corev1.ResourceMemory: resource.Quantity{Format: "200Mi"},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "GITKUBE_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const SVCName = "gitkubed"

func newSVC(namespace string, svcType corev1.ServiceType) corev1.Service {
	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SVCName,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "ssh",
					Port: 22,
					TargetPort: intstr.IntOrString{
						IntVal: 22,
					},
				},
			},
			Selector: map[string]string{"app": "gitkubed"},
			Type:     svcType,
		},
	}
}

func int2ptr(i int32) *int32 {
	return &i
}
