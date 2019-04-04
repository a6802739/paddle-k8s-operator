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

package paddlejob

import (
	"fmt"

	paddleresource "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Cluster is our interface to the Kubernetes cluster. It can inquiry
// the cluster's overall status and the status of a specific
// PaddlePaddle trainning job.  It can also create training jobs and
// replica.
//
// TODO(yi): The above functionalities are NOT logically related with
// each other.  I am not sure if it is a good idea to group them in
// this source file.
type Cluster struct {
	clientset *kubernetes.Clientset
}

// newCluster create a new instance of K8sCluster.
func newCluster(clientset *kubernetes.Clientset) *Cluster {
	return &Cluster{
		clientset: clientset,
	}
}

// GetTrainerJob gets the trainer job spec.
func (c Cluster) GetTrainerJob(job *paddleresource.PaddleJob) (*batchv1.Job, error) {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
}


// JobPods returns the number total desired pods and the number of
// running pods of a job.
func (c Cluster) JobPods(job *paddleresource.PaddleJob) (total, running, succeeded, pending int, err error) {
	if err != nil {
		return
	}
	// get pods of the job
	jobPods, err := c.clientset.CoreV1().
		Pods(job.ObjectMeta.Namespace).
		List(metav1.ListOptions{LabelSelector: "paddle-job=" + job.ObjectMeta.Name})
	for _, pod := range jobPods.Items {
		total++
		// pod.ObjectMeta.DeletionTimestamp means pod is terminating
		if pod.ObjectMeta.DeletionTimestamp == nil && pod.Status.Phase == v1.PodRunning {
			running++
		}
		if pod.ObjectMeta.DeletionTimestamp == nil && pod.Status.Phase == v1.PodPending {
			pending++
		}
		if pod.Status.Phase == v1.PodSucceeded {
			succeeded++
		}
	}
	return
}
