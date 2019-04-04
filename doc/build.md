# How to Build Operator Component

This article contains instructions of build paddle operator and Docker image so that the component can run in the Kubernetes cluster.

## Build Operator

```bash
glide install --strip-vendor
go build github.com/paddlepaddle/paddlejob/cmd/paddlejob
```

The above step will generate a binary file named `paddlejob` which should
run as a daemon process on the Kubernetes cluster.

## Build Operator Image

To build your own docker images, run the following command:

```bash
docker build -t yourRepoName/paddle-operator .
```

This command will take the `Dockerfile`, build the docker image and tag it as `yourRepoName/paddle-operator`

Now you want to push it to your docker hub so that Kubernetes cluster is able to pull and deploy it.

``` bash
docker push yourRepoName/paddle-operator
```
