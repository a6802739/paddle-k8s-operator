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

// Controller is responsible to watch resource type "PaddleJob"
// event and parse "PaddleJob" into several other resources like
// "Job" and "ReplicaSet".

// Controller will manage "PaddleJob" creation and destruction while
// PaddleJobSynced will monitor the cluster resources and training jobs.

// When controller starts, both event watching routine and resource
// monitoring should be started.

package paddlejob

import (
	"encoding/json"
	"sync"

	log "github.com/inconshreveable/log15"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/kubernetes/pkg/api"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	paddleJobClient "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned"
	paddleresource "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
)

// Controller for dispatching PaddleJob resource.
type Controller struct {
	client     *rest.RESTClient
	clientset  *kubernetes.Clientset
	paddleJobSynced *PaddleJobSynced
}

// New construct a new Controller struct
func New(c *rest.RESTClient, cs *kubernetes.Clientset) (*Controller, error) {
	cluster := newCluster(cs)
	as := newPaddleJobSynced(cluster)

	return &Controller{
		client:     c,
		clientset:  cs,
		paddleJobSynced: as,
	}, nil
}

// Run start to watch kubernetes events and do handlers.
func (c *Controller) Run(paddleJobClient paddleJobClient.Interface) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c.WatchPaddleJobs()
		wg.Done()
	}()
	go func() {
		c.paddleJobSynced.Run(c.clientset, paddleJobClient)
		wg.Done()
	}()
	wg.Wait()
}

// WatchPaddleJobs moinitors paddleJobs resources.
func (c *Controller) WatchPaddleJobs() {
	source := cache.NewListWatchFromClient(
		c.client,
		paddleresource.PaddleJobs,
		// TODO(helin): pass in namespace as an argument.
		api.NamespaceAll,
		fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&paddleresource.PaddleJob{},

		// TODO(helin): support resync. resync will eventually
		// happen even if the resyncPeriod parameter is set to
		// 0.

		// resyncPeriod: Every resyncPeriod, all resources in
		// the cache will retrigger events. Set to 0 to
		// disable the resync.
		0,

		// PaddleJob custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	informer.Run(make(chan struct{})) // A channel will never close.
}

func (c *Controller) onAdd(obj interface{}) {
	job := obj.(*paddleresource.PaddleJob)
	log.Debug("PaddleJob resource added", "name", job.ObjectMeta.Name)
	c.paddleJobSynced.OnAdd(job)

	// TODO(gongwb):open it when all are ready.
	// All-are-ready means:
	//  create trainjob from paddlectl
	//  scheduler can schedule trainjobs
	var parser DefaultJobParser
	p := parser.ParseToPserver(job)
	t := parser.ParseToTrainer(job)

	b, _ := json.MarshalIndent(p, "", "   ")
	log.Debug("create pserver:" + string(b))

	b, _ = json.MarshalIndent(t, "", "   ")
	log.Debug("create trainer-job:" + string(b))

	// create all resources
	_, err := c.clientset.ExtensionsV1beta1().ReplicaSets(p.ObjectMeta.Namespace).Create(p)
	if err != nil {
		log.Error("create pserver", "error", err)
	}

	_, err = c.clientset.BatchV1().Jobs(t.ObjectMeta.Namespace).Create(t)
	if err != nil {
		log.Error("create trainer", "error", err)
	}
}

func (c *Controller) onUpdate(oldObj, newObj interface{}) {
	oldjob := oldObj.(*paddleresource.PaddleJob)
	newjob := newObj.(*paddleresource.PaddleJob)
	log.Debug("PaddleJob resource updated", "old name", oldjob.ObjectMeta.Name, "new name", newjob.ObjectMeta.Name)
	c.paddleJobSynced.OnUpdate(newjob)
}

func (c *Controller) onDelete(obj interface{}) {
	job := obj.(*paddleresource.PaddleJob)
	log.Debug("PaddleJob resource deleted", "name", job.ObjectMeta.Name)
	c.paddleJobSynced.OnDel(job)
}
