#!/usr/bin/env bash
# File managed by pluginsync

# http://www.apache.org/licenses/LICENSE-2.0.txt
#
#
# Copyright 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

detect_go_dep() {
  [[ -f "${__proj_dir}/Godeps/Godeps.json" ]] && _dep='godep'
  [[ -f "${__proj_dir}/glide.yaml" ]] && _dep='glide'
  [[ -f "${__proj_dir}/vendor/vendor.json" ]] && _dep='govendor'
  _info "golang dependency tool: ${_dep}"
  echo "${_dep}"
}

install_go_dep() {
  local _dep=${_dep:=$(_detect_dep)}
  _info "ensuring ${_dep} is available"
  case $_dep in
    godep)
      _go_get github.com/tools/godep
      ;;
    glide)
      _go_get github.com/Masterminds/glide
      ;;
    govendor)
      _go_get github.com/kardianos/govendor
      ;;
  esac
}

restore_go_dep() {
  local _dep=${_dep:=$(_detect_dep)}
  _info "restoring dependency with ${_dep}"
  case $_dep in
    godep)
      (cd "${__proj_dir}" && godep restore)
      ;;
    glide)
      (cd "${__proj_dir}" && glide install)
      ;;
    govendor)
      (cd "${__proj_dir}" && govendor sync)
      ;;
  esac
}

_dep=$(detect_go_dep)
install_go_dep
restore_go_dep
