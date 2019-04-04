# Distributed Mnist Examples

This folder contains an example where mnist is trained.

The python script used to train linear regression with paddle fluid takes in several arguments that can be used to switch the distributed backends. The manifests to launch the distributed training of this dist_train file using the paddle operator. This folder contains manifest with example usage.

## Build Image

The default image name and tag is kubeflow/paddle-mnist-test:1.0.
> docker build -f Dockerfile -t kubeflow/paddle-mnist-test:1.0 ./
> docker push kubeflow/paddle-mnist-test:1.0

## Create the Paddle job

There are two methods to create paddle job. Details are as follows:
1. The below example uses the yaml file.
> kubectl create -f fluid_job_recognize_digits.yaml

2. Using [ksonnet tool](https://ksonnet.io/docs/tutorial/guestbook/) to create paddle job component.
