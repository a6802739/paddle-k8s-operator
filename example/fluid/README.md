# Deployment tips

1. Building image. If you want to create your own image we recommend using dockerhub. Each example has its own Dockerfile that we strongly advise to use. To build your custom image follow instruction on [TechRepublic](https://www.techrepublic.com/article/how-to-create-a-docker-image-and-push-it-to-docker-hub/).

2. To deploy your job we recommend using official kubeflow documentation. Each example has many parameters. Feel free to modify them based on ksonnet, e.g. image or number of GPUs.
