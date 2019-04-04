# Distributed Linear regression Examples

This folder contains an example where linear regression is trained.

The python script used to train linear regression with paddle fluid takes in several arguments that can be used to switch the distributed backends. The manifests to launch the distributed training of this dist_train file using the paddle operator. This folder contains manifest with example usage.

## Build Image

The default image name and tag is kubeflow/paddle-dist-train-test:1.0.
> docker build -f Dockerfile -t kubeflow/paddle-dist-train-test:1.0 ./
> docker push kubeflow/paddle-dist-train-test:1.0

## Create the Paddle job

There are two methods to create paddle job. Details are as follows:
1. The below example uses the yaml file.
> kubectl create -f fluid_job_dist_train.yaml

2. Using [ksonnet tool](https://ksonnet.io/docs/tutorial/guestbook/) to create paddle job component.
