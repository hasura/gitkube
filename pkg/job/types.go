package job

import (
	"github.com/hasura/gitkube/pkg/providers"
	"k8s.io/client-go/kubernetes"
)

type Status string

const (
	Created Status = "created"
	Failed  Status = "failed"
	Success Status = "success"
)

type JobSpec struct {
	Provider string
	Status   Status
	Output   string
	Remote   string
	Branch   string
	Uuid     string
}

type JobClient struct {
	kubeclientset *kubernetes.Clientset
}

type BuildSpec struct {
	Provider string
	Remote   string
	Branch   string
	PushSpec providers.PushSpec
}
