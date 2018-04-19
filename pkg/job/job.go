package job

import (
	"encoding/json"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	gitkubeNamespace string
)

func SetGitkubeNamespace(ns string) {
	if ns == "" {
		gitkubeNamespace = "kube-system"

	} else {
		gitkubeNamespace = ns
	}
}

func New(kubeclientset *kubernetes.Clientset) *JobClient {
	return &JobClient{
		kubeclientset: kubeclientset,
	}
}

func (jc *JobClient) Create(spec BuildSpec) (*JobSpec, error) {
	job := JobSpec{}
	job.Status = Created
	job.Provider = spec.Provider
	job.Remote = spec.Remote
	job.Branch = spec.Branch
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	job.Uuid = uuid.String()
	cm, err := jc.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Get("gitkube-jobs-conf", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(cm.Data) == 0 {
		cm.Data = make(map[string]string)
	}

	bytes, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	cm.Data[job.Uuid] = string(bytes)

	_, err = jc.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Update(cm)
	if err != nil {
		return nil, err
	}

	return &job, nil

}

func (jc *JobClient) Update(job *JobSpec) (*JobSpec, error) {
	cm, err := jc.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Get("gitkube-jobs-conf", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(*job)
	if err != nil {
		return nil, err
	}

	cm.Data[job.Uuid] = string(bytes)

	_, err = jc.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Update(cm)
	if err != nil {
		return nil, err
	}

	return job, nil

}
