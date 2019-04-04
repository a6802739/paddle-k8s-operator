#!/bin/bash
# ks pkg remove kubeflow/paddle-job
# ks delete default -c my-paddle-operator
# ks delete default -c my-paddle-job
# ks component rm my-paddle-job
# ks component rm my-paddle-operator
# ks pkg install kubeflow/paddle-job
# ks generate paddle-job my-paddle-job
# ks generate paddle-operator my-paddle-operator

JOB_NAME=my-paddle-job
ks param set ${JOB_NAME} image "ppl521/paddle-fluid-job:25.0"
ks param set ${JOB_NAME} port 7164
ks param set ${JOB_NAME} ports_num 1
ks param set ${JOB_NAME} port_num_for_sparse 1
ks param set ${JOB_NAME} trainer_limit_cpu "200m"
ks param set ${JOB_NAME} trainer_limit_mem "200Mi"
ks param set ${JOB_NAME} trainer_request_cpu "200m"
ks param set ${JOB_NAME} trainer_request_mem "200Mi"
ks param set ${JOB_NAME} trainer_max_instance 2
ks param set ${JOB_NAME} trainer_min_instance 2
ks param set ${JOB_NAME} trainer_gpu 0
ks param set ${JOB_NAME} entrypoint "python /workspace/recognize_digits.py train"
ks param set ${JOB_NAME} workspace "/workspace"
ks param set ${JOB_NAME} passes 50
ks param set ${JOB_NAME} pserver_limit_cpu "800m"
ks param set ${JOB_NAME} pserver_limit_mem "600Mi"
ks param set ${JOB_NAME} pserver_request_cpu "500m"
ks param set ${JOB_NAME} pserver_request_mem "200Mi"
ks param set ${JOB_NAME} pserver_max_instance 2
ks param set ${JOB_NAME} pserver_min_instance 1

ks show default -c my-paddle-job
ks apply default -c my-paddle-job
#ks apply default -c my-paddle-operator
