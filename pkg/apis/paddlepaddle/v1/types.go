package v1

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoapi "k8s.io/client-go/kubernetes/scheme"
)

const (
	// CRDKind is the kind of K8s CRD.
	CRDKind = "PaddleJob"
	// CRDKindPlural is the plural of CRDKind.
	CRDKindPlural = "paddlejobs"
	// CRDShortName is the short name of CRD.
	CRDShortName = "tj"
	// CRDGroup is the name of group.
	CRDGroup = "paddlepaddle.org"
	// CRDVersion is the version of CRD.
	CRDVersion = "v1"
	// TrainingJobs string for registration
	PaddleJobs = "PaddleJobs"
)

// CRDName returns name of crd
func CRDName() string {
	return fmt.Sprintf("%s.%s", CRDKindPlural, CRDGroup)
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=PaddleJob

// PaddleJob is a specification for a PaddleJob resource
type PaddleJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PaddleJobSpec   `json:"spec"`
	Status            PaddleJobStatus `json:"status"`
}

// PaddleJobSpec is the spec for a PaddleJob resource
type PaddleJobSpec struct {
	// General job attributes.
	Image string `json:"image,omitempty"`
	// If you want to use the hostnetwork instead of container network
	// portmanager is necessary. About portmanager, please refer to
	// https://github.com/PaddlePaddle/cloud/blob/develop/doc/hostnetwork/hostnetwork.md
	HostNetwork       bool                 `json:"host_network,omitempty"`
	Port              int                  `json:"port,omitempty"`
	PortsNum          int                  `json:"ports_num,omitempty"`
	PortsNumForSparse int                  `json:"ports_num_for_sparse,omitempty"`
	Passes            int                  `json:"passes,omitempty"`
	Volumes           []corev1.Volume      `json:"volumes"`
	VolumeMounts      []corev1.VolumeMount `json:"VolumeMounts"`
	NodeSelector      map[string]string    `json:"NodeSelector"`
	//TODO(m3ngyang) simplify the structure of sub-resource(mengyang)
	//PaddleJob components.
	Pserver PserverSpec `json:"pserver"`
	Trainer TrainerSpec `json:"trainer"`
}

// PserverSpec is the spec for pservers in the paddle job
type PserverSpec struct {
	MinInstance int                         `json:"min-instance"`
	MaxInstance int                         `json:"max-instance"`
	Resources   corev1.ResourceRequirements `json:"resources"`
	ReplicaSpec *v1beta1.ReplicaSet         `json:"replicaSpec"`
}

// TrainerSpec is the spec for trainers in the paddle job
type TrainerSpec struct {
	Entrypoint   string                      `json:"entrypoint"`
	Workspace    string                      `json:"workspace"`
	MinInstance  int                         `json:"min-instance"`
	MaxInstance  int                         `json:"max-instance"`
	Resources    corev1.ResourceRequirements `json:"resources"`
	ReplicaSpec  *batchv1.Job                `json:"replicaSpec"`
}

// PaddleJobPhase is the phase of PaddleJob
type PaddleJobPhase string

const (
	// PaddleJobPhaseNone is empty PaddleJobPhase.
	PaddleJobPhaseNone PaddleJobPhase = ""
	// PaddleJobPhaseCreating is creating PaddleJobPhase.
	PaddleJobPhaseCreating = "creating"
	// PaddleJobPhaseRunning is running PaddleJobPhase.
	PaddleJobPhaseRunning = "running"
	// PaddleJobPhaseSucceeded is succeeded PaddleJobPhase.
	PaddleJobPhaseSucceeded = "succeeded"
	// PaddleJobPhaseFailed is failed PaddleJobPhase.
	PaddleJobPhaseFailed = "failed"
)

// TrainingResourceType the type of PaddleJob resource, include PSERVER and TRAINER
type TrainingResourceType string

const (
	// Pserver is the pserver name of TrainingResourceType.
	Pserver TrainingResourceType = "PSERVER"
	// Trainer is the trainer name of TrainingResourceType.
	Trainer TrainingResourceType = "TRAINER"
)

// ResourceState is the state of a type of resource
type ResourceState string

const (
	// ResourceStateNone is the initial state of training job
	ResourceStateNone ResourceState = ""
	// ResourceStateStarting is the starting state of ResourceState.
	ResourceStateStarting = "starting"
	// ResourceStateRunning is the  running state of ResourceState.
	ResourceStateRunning = "running"
	// ResourceStateFailed is the failed state of ResourceState.
	ResourceStateFailed = "failed"
	// ResourceStateSucceeded is the succeeded state of ResourceState
	ResourceStateSucceeded = "succeeded"
)

// TrainingResourceStatus is the status of every resource
type TrainingResourceStatus struct {
	// TrainingResourceType the type of PaddleJob resource, include PSERVER and TRAINER
	TrainingResourceType `json:"training_resource_type"`
	// State is the state of a type of resource
	State ResourceState `json:"state"`
	// ResourceStates is the number of resource in different state
	ResourceStates map[ResourceState]int `json:"resource_states"`
}

// PaddleJobStatus is the status for a PaddleJob resource.
type PaddleJobStatus struct {
	// Phase is phase of PaddleJob
	Phase PaddleJobPhase `json:"phase"`
	// Reason is the reason of job phase failed
	Reason string `json:"reason"`
	// ReplicaStatuses is detail status of resources
	// TODO(ZhengQi): should we only considered trainer job now?
	ReplicaStatuses []*TrainingResourceStatus `json:"replica_statuses"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=trainingjobs

// PaddleJobList is a list of PaddleJob resources
type PaddleJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	// Items means the list of paddle job/PaddleJob
	Items []PaddleJob `json:"items"`
}

func RegisterResource(config *rest.Config, resourceType, resourceListType runtime.Object) *rest.Config {
	groupversion := schema.GroupVersion{
		Group:   "paddlepaddle.org",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: clientgoapi.Codecs}

	clientgoapi.Scheme.AddKnownTypes(
		groupversion,
		resourceType,
		resourceListType,
		&corev1.ListOptions{},
		&corev1.DeleteOptions{},
	)

	return config
}
