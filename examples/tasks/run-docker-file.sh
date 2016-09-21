#!/bin/bash

set -e
set -u
set -o pipefail

# get the directory the script exists in
__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(cd $__dir && cd ../../ && pwd)"
__proj_name="$(basename $__proj_dir)"

export PLUGIN_SRC="${__proj_dir}"
export PLUGIN_DEST="/${__proj_name}"

# source the common bash script 
. "${__proj_dir}/scripts/common.sh"

# verifies dependencies and starts bind
. "${__proj_dir}/examples/tasks/.setup.sh"

# dockerception.sh will create the Snap container and run the $RUN_SCRIPT
cd "${__proj_dir}/examples/tasks" && docker-compose exec docker sh -c "${PLUGIN_DEST}/examples/tasks/run-dockerception.sh"