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
	"time"

	// TODO(typhoonzero): this package still depends on k8s API, try to remove this.
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
	log "github.com/inconshreveable/log15"
	paddleresource "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
	paddleJobClient "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned"
	"github.com/paddlepaddle/paddlejob/pkg/updater"
)

const (
	defaultLoopDur = time.Second * 5
)

type job struct {
	Config     *paddleresource.PaddleJob
	TrainerJob *batchv1.Job
}

// PaddleJobSynced launches the training jobs.
type PaddleJobSynced struct {
	ticker         *time.Ticker
	cluster        *Cluster
	jobs           map[string]*job
	eventCh        chan event
}

// newPaddleJobSynced creates a new PaddleJobSynced.
func newPaddleJobSynced(cluster *Cluster, options ...func(*PaddleJobSynced)) *PaddleJobSynced {
	c := &PaddleJobSynced{
		cluster:        cluster,
		ticker:         time.NewTicker(defaultLoopDur),
		jobs:           make(map[string]*job),
		eventCh:        make(chan event),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

type jobs []*job

type eventType int

const (
	add eventType = iota
	del
	update
)

type event struct {
	Type eventType
	Job  *paddleresource.PaddleJob
}

// OnAdd notifies the paddleJobSynced that a job has been added.
func (a *PaddleJobSynced) OnAdd(PaddleJob *paddleresource.PaddleJob) {
	a.eventCh <- event{Type: add, Job: PaddleJob}
}

// OnDel notifies the paddleJobSynced that a job has been deleted.
func (a *PaddleJobSynced) OnDel(PaddleJob *paddleresource.PaddleJob) {
	a.eventCh <- event{Type: del, Job: PaddleJob}
}

// OnUpdate notifies the paddleJobSynced that a job has been deleted.
func (a *PaddleJobSynced) OnUpdate(PaddleJob *paddleresource.PaddleJob) {
	a.eventCh <- event{Type: update, Job: PaddleJob}
}

// updateJobList updates the data structure a.jobs according to
// received events about the PaddleJob resource.  It returns true if
// the controller need to do some scheduling work.  If it returns
// false, the controller could simply go on monitoring other
// events.
func (a *PaddleJobSynced) updateJobList(evt event) bool {
	log.Debug("monitor received event", "event", evt)
	switch evt.Type {
	case add, update:
		j := &job{Config: evt.Job}
		a.jobs[evt.Job.ObjectMeta.Name] = j
		if a.tryToRetrieveTrainerJobInPaddleJob(evt.Job.ObjectMeta.Name, j) != nil {
			return false
		}
	case del:
		// TODO(helin): delete all created resources (e.g.,
		// trainer Job, pserver Replica Set) when we schedules
		// the resources.
		delete(a.jobs, evt.Job.ObjectMeta.Name)
	default:
		log.Error("unrecognized event", "event", evt)
	}

	return true
}

// findSucceededJob returns true if there is at least one training job
// whose all pods are Succeeded.
func (a *PaddleJobSynced) findSucceededJob(clientset *kubernetes.Clientset, paddleJobClient paddleJobClient.Interface) bool {
	for jobName, job := range a.jobs {
		if a.tryToRetrieveTrainerJobInPaddleJob(jobName, job) != nil {
			continue
		}
		total, _, succeeded, _, err := a.cluster.JobPods(job.Config)
		if err != nil {
			log.Error("check if job is running failed", "error", err)
			continue
		}

		if total == succeeded {
			if _, err := updater.NewUpdater(job.Config, clientset, paddleJobClient); err != nil{
				return false
			}
			return true
		}
	}
	return false
}


func (a *PaddleJobSynced) tryToRetrieveTrainerJobInPaddleJob(jobName string, job *job) error {
	// TODO(helin): Because we wrote the conversion from
	// PaddleJob into ReplicaSet and Job in paddlelcloud instead
	// of controller (please refer to above TODO comment for
	// details), we might suffer from the problem that when the
	// controller calls cluster.GetTrainerJob, it doesn't
	// return TrainerJob before the conversion is done at the
	// paddlecloud.  After we fix this problem, we can remove the
	// following call to cluster.GetTrainerJob and keep the one in
	// updateJobLists.
	if job.TrainerJob == nil {
		tj, err := a.cluster.GetTrainerJob(job.Config)
		if err != nil {
			log.Error(
				"Error getting the trainer k8s job, will sync later.",
				"name", job.Config.ObjectMeta.Name,
				"error", err,
			)
			return err
		}
		job.TrainerJob = tj
	}
	return nil
}

// Run monitors the cluster resources and training jobs in a loop.
func (a *PaddleJobSynced) Run(clientset *kubernetes.Clientset, paddleJobClient paddleJobClient.Interface) {
	for {
		select {
		case <-a.ticker.C:
		case evt := <-a.eventCh:
			if !a.updateJobList(evt) {
				continue // If nothing important, go on looping.
			}
		}
		a.findSucceededJob(clientset, paddleJobClient)
	}
}
