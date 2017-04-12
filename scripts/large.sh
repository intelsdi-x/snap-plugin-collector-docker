#!/usr/bin/env bash
# File managed by pluginsync

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"
__proj_name="$(basename "$__proj_dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_verify_docker() {
  type -p docker > /dev/null 2>&1 || _error "docker needs to be installed"
  type -p docker-compose > /dev/null 2>&1 || _error "docker-compose needs to be installed"
  docker version >/dev/null 2>&1 || _error "docker needs to be configured/running"
}

_verify_docker

[[ -f "${__proj_dir}/build/linux/x86_64/${__proj_name}" ]] || (cd "${__proj_dir}" && make)

SNAP_VERSION=${SNAP_VERSION:-"latest"}
OS=${OS:-"alpine"}
PLUGIN_PATH=${PLUGIN_PATH:-"${__proj_dir}"}
DEMO=${DEMO:-"false"}
TASK=${TASK:-""}

if [[ ${DEBUG:-} == "true" ]]; then
  cmd="cd /plugin/scripts && rescue rspec ./test/*_spec.rb"
else
  cmd="cd /plugin/scripts && rspec ./test/*_spec.rb"
fi

_info "running large test"
docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${__proj_dir}":/plugin -e DEMO="${DEMO}" -e TASK="${TASK}" -e PLUGIN_PATH="${PLUGIN_PATH}" -e SNAP_VERSION="${SNAP_VERSION}" -e OS="${OS}" -ti intelsdi/serverspec:alpine /bin/sh -c "${cmd}"
