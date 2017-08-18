#!/bin/bash

#http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
#Copyright 2016 Intel Corporation
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

. "${__dir}/common.sh"

_info "running the example ${__proj_dir}/examples/docker-file.sh"
export PLUGIN_PATH="/etc/snap/path"
source "${__proj_dir}/examples/docker-file.sh"

_debug "sleeping for 10 seconds so the task can do some work"
sleep 20

# begin assertions
return_code=0
echo -n "[task is running] "
task_list=$(snaptel task list | tail -1)
if echo $task_list | grep -q Running; then
    echo "ok"
else
    echo "not ok"
    return_code=-1
fi

echo -n "[task is hitting] "
if [ $(echo $task_list | awk '{print $4}') -gt 0 ]; then
    echo "ok"
else
    _debug $task_list
    echo "not ok"
    return_code=-1
fi

echo -n "[task has no errors] "
if [ $(echo $task_list | awk '{print $6}') -eq 0 ]; then
    echo "ok"
else
    echo "not ok"
    return_code=-1
fi

exit $return_code
