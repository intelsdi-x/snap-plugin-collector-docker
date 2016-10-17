<!--
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-collector-docker.svg?branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-collector-docker)

# snap collector plugin - Docker

This plugin collects runtime metrics from Docker containers and its host machine. It gathers information about resource usage and performance characteristics. 

It's used in the [Snap framework](http://github.com/intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license-and-authors)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need Docker Engine installed. Visit [Install Docker Engine](https://docs.docker.com/engine/installation/) for detailed instructions how to do it.

### Operating systems
* Linux/amd64
* Darwin/amd64 (needs [docker-machine](https://docs.docker.com/v1.8/installation/mac/))

### Installation

You can get the pre-built binaries for your OS and architecture at snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-docker
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-docker.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
It may take a while to pull dependencies if you haven't had them already.

This builds the plugin in `/build/`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Load the plugin and create a task, see example in [Examples](https://github.com/intelsdi-x/snap-plugin-collector-docker/blob/master/README.md#examples).

#### Configuration parameters
* Set environment variable `PROCFS_MOUNT` to point to the path where proc of host is mounted.

## Documentation
There are a number of other resources you can review to learn to use this plugin:
* [Docker documentation](https://docs.docker.com/)
* [Docker runtime metrics](https://docs.docker.com/v1.9/engine/articles/runmetrics/)

Notice, that this plugin using default docker server endpoint `unix:///var/run/docker.sock` to communicate with docker deamon.
However, adding support for custom endpoints is on Roadmap.

Client instance ready for communication with the given
// server endpoint. It will use the latest remote API version available in the
// server.
### Collected Metrics

The list of collected metrics is described in [METRICS.md](METRICS.md).

### Examples
Similar to dream levels in the movie _Inception_, we have different levels of examples:
* LEVEL 0: Snap running on your system (Linux only).
* LEVEL 1: Snap runs in a container.
* LEVEL 2: Snap runs in a docker-in-docker container.

For the sake of ease-of-use, these examples are presented in reverse order.

#### Run example in a docker-in-docker container
```
./examples/tasks/run-docker-file.sh
```

#### Run example in a docker container (Linux or Darwin only)
```
./examples/tasks/run-dockerception.sh
```

#### Run example on your Linux system

Check if there is some running docker container(s):

```
$ docker ps

CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
7720efd76bb8        ubuntu              "/bin/bash"         35 minutes ago      Up 35 minutes                           prickly_spence
ad5221e8ae73        ubuntu              "/bin/bash"         36 minutes ago      Up 36 minutes                     		suspicious_mirzakhani
```


In one terminal window, start the Snap daemon (in this case with logging set to 1 and trust disabled):
```
$ snapd -l 1 -t 0
```

In another terminal window:
Load snap-plugin-collector-docker plugin:
```
$ snapctl plugin load build/linux/x86_64/snap-plugin-collector-docker
```

Get file plugin for publishing, appropriate for Linux or Darwin:
```
$ wget  http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
```
or
```
$ wget  http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/darwin/x86_64/snap-plugin-publisher-file
```

Load file plugin for publishing:
```
$ snapctl plugin load snap-plugin-publisher-file
```

Another terminal window, you can list all of available metrics:
```
$ snapctl metric list
```

Create task manifest for writing to a file. You can also use an asterisk - see [`examples/tasks/docker-file.json`](examples/tasks/docker-file.json) for which all exposed metrics for all containers will be collected:
```json
{
  "version": 1,
  "schedule": {
    "type": "simple",
    "interval": "1s"
  },
  "workflow": {
    "collect": {
      "metrics": {
        "/intel/docker/*": {}
      },
      "config": {},
      "publish": [
        {
          "plugin_name": "file",
          "config": {
            "file": "/tmp/snap-docker-file.log"
          }
        }
      ]
    }
  }
}
```
Create a task by the following command:
```
$ snapctl task create -t examples/tasks/docker-file.json

Using Task Manifest to create task
Task created
ID: da941d6f-e137-4ef8-97cd-e2b73a8559fb
Name: Task-da941d6f-e137-4ef8-97cd-e2b73a8559fb
State: Running
```
See  output from snapctl task watch <task_id>

(notice, that below only the fragment of task watcher output has been presented)

```
$ snapctl task watch da941d6f-e137-4ef8-97cd-e2b73a8559fb

Watching Task (da941d6f-e137-4ef8-97cd-e2b73a8559fb):

NAMESPACE                                                                    DATA      		TIMESTAMP

/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/total_usage           2.146646e+07       2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/usage_in_kernelmode   1e+07              2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/usage_in_usermode     0                  2016-06-21 12:44:09.552146203 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/total_usage           2.146646e+07       2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/usage_in_kernelmode   1e+07              2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/usage_in_usermode     0                  2016-06-21 12:44:09.552146203 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/total_usage                   2.88984998661e+12  2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/usage_in_kernelmode           6.38e+11            2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/usage_in_usermode             9.4397e+11          2016-06-21 12:44:09.552146203 +0200 CEST

```
(Keys `ctrl+c` terminate task watcher)

These data are published to file and stored there (in this example in `/tmp/snap-docker-file.log`).

### Roadmap
This plugin is in active development. As we lauch this plugin, we have a few item in mind for the next release:
- [ ] Support for custom docker endpoints different that default `unix:///var/run/docker.sock`
As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-docker/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-docker/pulls).

## Community Support
This repository is one of **many** plugins in the **snap framework**: a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author:       [Marcin Krolik](https://github.com/marcin-krolik)
* Co-authors:   [Izabella Raulin](https://github.com/IzabellaRaulin), [Marcin Olszewski](https://github.com/marcintao)

**Thank you!** Your contribution is incredibly important to us.
