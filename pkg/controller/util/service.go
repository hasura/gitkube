package util

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// GetExternalIP gets the name or IP of (gitkubed) service
// Returns error for unsupported Service types
func GetExternalIP(kubeclientset *kubernetes.Clientset, service *corev1.Service) (string, error) {
	switch service.Spec.Type {
	case corev1.ServiceTypeClusterIP:
		return "", fmt.Errorf("gitkube service type %s cannot be accessed from outside cluster. If this was intended, add remote manually",
			corev1.ServiceTypeClusterIP)

	case corev1.ServiceTypeExternalName:
		return service.Spec.ExternalName, nil

	case corev1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) == 0 {
			return "", fmt.Errorf("gitkube service of type %s has no available IP/hostnames", corev1.ServiceTypeLoadBalancer)
		}

		loadbalancerIPOrName := GetLoadBalancerIPOrName(service.Status.LoadBalancer.Ingress[0])

		if loadbalancerIPOrName == "" {
			return "", fmt.Errorf("gitkube service of type %s has unknown IP/hostname", corev1.ServiceTypeLoadBalancer)
		}

		return loadbalancerIPOrName, nil

	case corev1.ServiceTypeNodePort:
		return "", fmt.Errorf("manually configure remote for gitkube service of type %s. E.g.: ssh://<namespace>-<remote-name>@<any-node-ip>:<node-port>/~/git/<namespace>-<remote-name>",
			corev1.ServiceTypeNodePort)
	default:
		return "", fmt.Errorf("unknown gitkubed service type %s. Is gitkubed service exposed?", service.Spec.Type)
	}

	return "", fmt.Errorf("could not process gitkube service ip")

}

// GetLoadBalancerIPOrName gets the name or IP from LoadBalancerIngress resource
func GetLoadBalancerIPOrName(ingress corev1.LoadBalancerIngress) string {
	if ingress.IP != "" {
		return ingress.IP
	} else if ingress.Hostname != "" {
		return ingress.Hostname
	} else {
		return ""
	}
}
