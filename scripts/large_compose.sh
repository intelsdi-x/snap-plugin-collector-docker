#!/bin/bash 

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename $__proj_dir)"

. "${__dir}/common.sh"

# NOTE: these variables control the docker-compose image.
export PLUGIN_SRC="${__proj_dir}"
export PROJ_NAME="${__proj_name}"
export LOG_LEVEL="${LOG_LEVEL:-"7"}"
export PLUGIN_DEST="/${__proj_name}"

TEST_TYPE="${TEST_TYPE:-"large"}"

docker_folder="${__proj_dir}/examples"

_docker_project () {
  (cd "${docker_folder}" && "$@")
}

_info "docker folder : $docker_folder"

_debug "running docker compose images"
_docker_project docker-compose up -d 
_debug "running test: ${TEST_TYPE}"

set +e

RUN_TESTS="\"${PLUGIN_DEST}/scripts/large_tests.sh\""
_docker_project docker-compose exec docker sh -c "export LOG_LEVEL=$LOG_LEVEL; export RUN_SCRIPT=$RUN_TESTS ; /${__proj_name}/examples/run-dockerception.sh"
test_res=$?
set -e
echo "exit code from large_compose $test_res"
_debug "stopping and removing containers"
_docker_project docker-compose down

exit $test_res