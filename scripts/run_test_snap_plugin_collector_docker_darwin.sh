#!/bin/bash -e

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

docker build -t intelsdi-x/snap-plugin-collector-docker -f snap-plugin-collector-docker/scripts/Dockerfile .
docker run -it -v /proc:/hostproc -v /sys/fs/cgroup:/sys/fs/cgroup  -v /var/run/docker.sock:/var/run/docker.sock -v ./static-docker:/usr/bin/docker intelsdi-x/snap-plugin-collector-docker bash
