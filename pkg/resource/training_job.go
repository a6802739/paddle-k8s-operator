// Copyright 2019 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoapi "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// PaddleJobs string for registration
const PaddleJobs = "PaddleJobs"

// A PaddleJob is a Kubernetes resource, it describes a PaddlePaddle
// training job.  As a Kubernetes resource,
//
//  - Its content must follow the Kubernetes resource definition convention.
//  - It must be a Go struct with JSON tags.
//  - It must implement the deepcopy interface.
//
// To start a PadldePaddle training job,
//
// (1) The user uses the paddlecloud command line tool, which sends
// the command line arguments to the paddlecloud HTTP server.
//
// (2) The paddlecloud server converts the command line arguments into
// a PaddleJob resource and sends it to the Kubernetes API server.
//
//
// (3) the controller, which moinitors events about the
// PaddleJob resource accepted by the Kubernetes API server,
// converts the PaddleJob resource into the following Kubernetes
// resources:
//
//   (3.1) a ReplicaSet of the parameter server proceses
//   (3.2)  a Job of trainer processes
//
// (4) some default controllers provided by Kubernetes monitors events
// about ReplicaSet and Job creates and maintains the Pods.
//
// An example PaddleJob instance:
/*
apiVersion: paddlepaddle.org/v1
kind: PaddleJob
metadata:
	name: job-1
spec:
	image: "paddlepaddle/paddlecloud-job"
	port: 7164
	ports_num: 1
	ports_num_for_sparse: 1
	fault_tolerant: true
	imagePullSecrets:
		name: myregistrykey
	hostNetwork: true
	trainer:
		entrypoint: "python train.py"
		workspace: "/home/job-1/"
		min-instance: 3
		max-instance: 6
		resources:
			limits:
				alpha.kubernetes.io/nvidia-gpu: 1
				cpu: "800m"
				memory: "1Gi"
			requests:
				cpu: "500m"
				memory: "600Mi"
	pserver:
		min-instance: 3
		max-instance: 3
		resources:
			limits:
				cpu: "800m"
				memory: "1Gi"
			requests:
				cpu: "500m"
				memory: "600Mi"
*/
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PaddleJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PaddleJobSpec   `json:"spec"`
	Status            PaddleJobStatus `json:"status,omitempty"`
}

// PaddleJobSpec defination
// +k8s:deepcopy-gen=true
type PaddleJobSpec struct {
	// General job attributes.
	Image             string                    `json:"image,omitempty"`
	Port              int                       `json:"port,omitempty"`
	PortsNum          int                       `json:"ports_num,omitempty"`
	PortsNumForSparse int                       `json:"ports_num_for_sparse,omitempty"`
	Passes            int                       `json:"passes,omitempty"`
	Volumes           []v1.Volume               `json:"volumes"`
	VolumeMounts      []v1.VolumeMount          `json:"VolumeMounts"`
	ImagePullSecrets  []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	HostNetwork       bool                      `josn:"hostNetwork,omitempty"`
	// Job components.
	Trainer TrainerSpec `json:"trainer"`
	Pserver PserverSpec `json:"pserver"`
}

// TrainerSpec defination
// +k8s:deepcopy-gen=true
type TrainerSpec struct {
	Entrypoint  string                  `json:"entrypoint"`
	Workspace   string                  `json:"workspace"`
	MinInstance int                     `json:"min-instance"`
	MaxInstance int                     `json:"max-instance"`
	Resources   v1.ResourceRequirements `json:"resources"`
}

// PserverSpec defination
// +k8s:deepcopy-gen=true
type PserverSpec struct {
	MinInstance int                     `json:"min-instance"`
	MaxInstance int                     `json:"max-instance"`
	Resources   v1.ResourceRequirements `json:"resources"`
}

// PaddleJobStatus defination
// +k8s:deepcopy-gen=true
type PaddleJobStatus struct {
	State   PaddleJobState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

// PaddleJobState defination
type PaddleJobState string

// PaddleJobState consts
const (
	StateCreated PaddleJobState = "Created"
	StateRunning PaddleJobState = "Running"
	StateFailed  PaddleJobState = "Failed"
	StateSucceed PaddleJobState = "Succeed"
)

// PaddleJobList defination
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PaddleJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []PaddleJob `json:"items"`
}

// GPU convert Resource Limit Quantity to int
func (s *PaddleJob) GPU() int {
	q := s.Spec.Trainer.Resources.Limits.NvidiaGPU()
	gpu, ok := q.AsInt64()
	if !ok {
		// FIXME: treat errors
		gpu = 0
	}
	return int(gpu)
}

// NeedGPU returns true if the job need GPU resource to run.
func (s *PaddleJob) NeedGPU() bool {
	return s.GPU() > 0
}

func (s *PaddleJob) String() string {
	b, _ := json.MarshalIndent(s, "", "   ")
	return string(b[:])
}

// RegisterResource registers a resource type and the corresponding
// resource list type to the local Kubernetes runtime under group
// version "paddlepaddle.org", so the runtime could encode/decode this
// Go type.  It also change config.GroupVersion to "paddlepaddle.org".
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
		&v1.ListOptions{},
		&v1.DeleteOptions{},
	)

	return config
}
