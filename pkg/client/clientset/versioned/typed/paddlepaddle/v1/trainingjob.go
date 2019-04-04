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

package v1

import (
	v1 "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
	scheme "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PaddleJobsGetter has a method to return a PaddleJobInterface.
// A group's client should implement this interface.
type PaddleJobsGetter interface {
	PaddleJobs(namespace string) PaddleJobInterface
}

// PaddleJobInterface has methods to work with PaddleJob resources.
type PaddleJobInterface interface {
	Create(*v1.PaddleJob) (*v1.PaddleJob, error)
	Update(*v1.PaddleJob) (*v1.PaddleJob, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.PaddleJob, error)
	List(opts meta_v1.ListOptions) (*v1.PaddleJobList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PaddleJob, err error)
	PaddleJobExpansion
}

// trainingJobs implements PaddleJobInterface
type trainingJobs struct {
	client rest.Interface
	ns     string
}

// newPaddleJobs returns a PaddleJobs
func newPaddleJobs(c *PaddlepaddleV1Client, namespace string) *trainingJobs {
	return &trainingJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the PaddleJob, and returns the corresponding PaddleJob object, and an error if there is any.
func (c *trainingJobs) Get(name string, options meta_v1.GetOptions) (result *v1.PaddleJob, err error) {
	result = &v1.PaddleJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PaddleJobs that match those selectors.
func (c *trainingJobs) List(opts meta_v1.ListOptions) (result *v1.PaddleJobList, err error) {
	result = &v1.PaddleJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested trainingJobs.
func (c *trainingJobs) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a PaddleJob and creates it.  Returns the server's representation of the PaddleJob, and an error, if there is any.
func (c *trainingJobs) Create(PaddleJob *v1.PaddleJob) (result *v1.PaddleJob, err error) {
	result = &v1.PaddleJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("trainingjobs").
		Body(PaddleJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a PaddleJob and updates it. Returns the server's representation of the PaddleJob, and an error, if there is any.
func (c *trainingJobs) Update(PaddleJob *v1.PaddleJob) (result *v1.PaddleJob, err error) {
	result = &v1.PaddleJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(PaddleJob.Name).
		Body(PaddleJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the PaddleJob and deletes it. Returns an error if one occurs.
func (c *trainingJobs) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *trainingJobs) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("trainingjobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched PaddleJob.
func (c *trainingJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PaddleJob, err error) {
	result = &v1.PaddleJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("trainingjobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
