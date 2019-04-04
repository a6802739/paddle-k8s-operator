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

package fake

import (
	v1 "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned/typed/paddlepaddle/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakePaddlepaddleV1 struct {
	*testing.Fake
}

func (c *FakePaddlepaddleV1) PaddleJobs(namespace string) v1.PaddleJobInterface {
	return &FakePaddleJobs{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakePaddlepaddleV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
