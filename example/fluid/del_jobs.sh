#!/bin/bash
function delete_job() {
  jobname=$1
  if [[ "$jobname" == "" ]]; then
      echo "Usage: sh del_jobs.sh [all|<job-name>]"
      exit 0
  fi
  kubectl delete PaddleJob $jobname
  kubectl delete job $jobname-trainer
  kubectl delete rs $jobname-pserver
}

function delete_all() {
  jobs=$(kubectl get PaddleJob | tail -n +2 | awk '{print $1}')
  for job in ${jobs[@]}
  do
    delete_job $job
  done
}

case "$1" in
    all)
      delete_all
      ;;
    *)
      delete_job $1
      ;;
esac
