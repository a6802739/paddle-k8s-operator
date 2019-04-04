# Kubernetes Custom Resource and Operator for PaddleJob

## Overview

This repository contains the specification and implementation of PaddleJob custom resource definition. Using this custom resource, users can create and manage Paddle Fluid jobs like other built-in resources in Kubernetes.

## Prerequisites
+ Kubernetes >= 1.8
+ kubectl

## What is PaddleJob?
PaddleJob is a Kubernetes custom resource that you can use to run Paddle training jobs on Kubernetes. The Kubeflow implementation of PaddleJob is in [paddle-job](kubeflow paddle-job component).

A PaddleJob is a resource with a simple YAML representation illustrated below.

```
apiVersion: paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: paddlejob
  namespace: testspace
spec:
  image: "<Your-docker-repo>/fluid_job_train_test:1.0"
  port: 7164
  ports_num: 1
  ports_num_for_sparse: 1
  mountPath: "/home/work/namespace/"
  pserver:
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  trainer:
    entrypoint: "python /home/job-1/train.py"
    workspace: "/home/job-1/"
    passes: 10
    min-instance: 2
    max-instance: 6
    resources:
      limits:
        cpu: "200m"
        memory: "200Mi"
      requests:
        cpu: "200m"
        memory: "200Mi"
```
## Installing Paddle Operator

There are two methods to install paddle operator:

1. Using Ksonnet to install

Before install operator, please refer to the installation instructions in the [Kubeflow user guide]( https://www.kubeflow.org/docs/started/getting-started/) to deploy kubeflow to your cluster. This installs paddlejob CRD and paddle-operator controller to manage the lifecycle of Paddle jobs.

> cd ${KSONNET_APP}

> ks pkg install kubeflow/paddle-job

> ks generate paddle-operator my-paddle-operator

> ks apply ${ENVIRONMENT} -c my-paddle-operator

2. Using Yaml file to install

Alternatively, you can deploy the operator with default settings without using ksonnet by running the following from the repo:
> kubectl create -f manifests/

## Creating a Paddle Job

There are two methods to create paddle jobs. Details are as follows:

1. Create paddle jobs using the kubeflow component.

**Note:**  We treat each paddle job as a component in your APP.
You can create PaddleJob by using [ksonnet tool](https://ksonnet.io/docs/tutorial/guestbook/) to create kubeflow component, and then use ksonnet to deploy PaddleJob to k8s to start training.

Kubeflow ships with a [ksonnet prototype](https://ksonnet.io/docs/concepts/#prototype) suitable for running the paddle job demo.
You can also use this prototype to generate a component which you can then customize for your jobs.

Create the component:

> JOB_NAME=my-paddle-job

> ks init ${CNN_JOB_NAME}

> cd ${CNN_JOB_NAME}

> ks registry add kubeflow-git <Your kubeflow repo>/kubeflow

> ks pkg install kubeflow-git/paddle-job

Run the generate command:
> ks generate paddle-job ${JOB_NAME}

Submit it:
> ks apply ${ENVIRONMENT} -c ${JOB_NAME}

2. You can create Paddle Job by defining a PaddleJob config file. See the manifests for the dist train example. You may change the config file based on your requirements.

> cat example/fluid/dist_train/fluid_job_dist_train.yaml

Deploy the PaddleJob resource to start training:
> kubectl create -f fliud_job_dist_train.yaml

You should now be able to see the created pods matching the specified number of replicas.
> kubectl get pods -l paddle-job-name=${JOB_NAME}

## Monitoring a Paddle Job
> kubectl get -o yaml PaddleJob ${JOB_NAME}
