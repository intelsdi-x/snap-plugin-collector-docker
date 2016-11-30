#!/bin/sh

set -e
set -u
set -o pipefail

# PROCFS_MOUNT is the path where /proc will be mounted on the Snap container
PROCFS_MOUNT="${PROCFS_MOUNT:-"/proc_host"}"
SNAP_VERSION="${SNAP_VERSION:-"latest"}"

# define the default plugin folder locations
PLUGIN_SRC="${PLUGIN_SRC:-"$(cd "$(dirname "$0")"/../../ && pwd)"}"
PLUGIN_DEST="${PLUGIN_DEST:-$PLUGIN_SRC}"

# docker-file.sh will download plugins, starts snap, load plugins and start a task
DEFAULT_SCRIPT="${PLUGIN_DEST}/examples/tasks/docker-file.sh && printf \"\n\nhint: type 'snaptel task list'\ntype 'exit' when your done\n\n\" && bash"
RUN_SCRIPT="export SNAP_VERSION=${SNAP_VERSION} && export PROCFS_MOUNT=${PROCFS_MOUNT} && ${RUN_SCRIPT:-$DEFAULT_SCRIPT}"


docker run -i --name dockerception -v /proc:${PROCFS_MOUNT} -v /var/lib/docker:/var/lib/docker -v /sys/fs/cgroup:/sys/fs/cgroup  -v /var/run/docker.sock:/var/run/docker.sock -v ${PLUGIN_SRC}:${PLUGIN_DEST} intelsdi/snap:alpine_test bash -c "$RUN_SCRIPT"
docker rm dockerception