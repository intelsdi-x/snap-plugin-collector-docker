#!/bin/bash -x

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2015 Intel Corporation
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename $__proj_dir)"
__deployment_file="$__dir/config/docker-deployment.yml"
__deployment_name="docker-deployment"

. "${__dir}/common.sh"

_debug "__dir ${__dir}"
_debug "__proj_dir ${__proj_dir}"
_debug "__proj_name ${__proj_name}"

_debug "start k8 deployment $__deployment_file"
kubectl create -f $__deployment_file
while ! [ "$(kubectl get po --no-headers | grep $__deployment_name | grep Running | awk '{print $2}')" = "1/1" ]; do
    kubectl get po --no-headers | grep $__deployment_name | grep CrashLoopBackOff && echo 'container failed' && exit 1
    echo 'waiting for pod to come up'
    sleep 5
done
_debug "copying the src into the runner"
kubectl exec $(kubectl get po --no-headers | grep $__deployment_name | grep Running | awk '{print $1}') -c main -i  -- mkdir /src
tar c  . | kubectl exec $(kubectl get po --no-headers | grep $__deployment_name | grep Running | awk '{print $1}') -c main -i  --  tar -x -C /src

set +e
_debug "running tests through the runner"
kubectl exec $(kubectl get po --no-headers | grep $__deployment_name | grep Running | awk '{print $1}') -c main -i -- /src/examples/tasks/run-dockerception.sh
test_res=$?
set -e
_debug "exit code $test_res"
_debug "removing k8 deployment"
kubectl delete -f $__deployment_file
exit $test_res

