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


package main

import (
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/paddlepaddle/paddlejob/pkg"
	paddleresource "github.com/paddlepaddle/paddlejob/pkg/apis/paddlepaddle/v1"
	paddleJobClient "github.com/paddlepaddle/paddlejob/pkg/client/clientset/versioned"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	var cfg *rest.Config
	if *kubeconfig != "" {
		cfg, _ = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		cfg, _ = rest.InClusterConfig()
	}

	paddleresource.RegisterResource(cfg, &paddleresource.PaddleJob{}, &paddleresource.PaddleJobList{})

	clientset, _ := kubernetes.NewForConfig(cfg)

	client, _ := rest.RESTClientFor(cfg)

	paddleJobClient, _ := paddleJobClient.NewForConfig(cfg)

	controller, _ := paddlejob.New(client, clientset)

	controller.Run(paddleJobClient)
}
