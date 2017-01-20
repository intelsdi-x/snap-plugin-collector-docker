[![Build Status](https://travis-ci.org/intelsdi-x/snap-plugin-collector-docker.svg?branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-collector-docker)
# Snap collector plugin - Docker

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
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need Docker Engine installed. Visit [Install Docker Engine](https://docs.docker.com/engine/installation/) for detailed instructions how to do it.
Plugin was tested against Docker version 1.12.3.

### Operating systems
* Linux/amd64
* Darwin/amd64 (needs [docker-machine](https://docs.docker.com/v1.8/installation/mac/))

### Installation
#### Download the plugin binary:

You can get the pre-built binaries for your OS and architecture from the plugin's [GitHub Releases](https://github.com/intelsdi-x/snap-plugin-collector-docker/releases) page. Download the plugin from the latest release and load it into `snapteld` (`/opt/snap/plugins` is the default location for Snap packages).

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-docker
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-docker.git
```

Build the Snap docker plugin by running make within the cloned repo:
```
$ make
```
It may take a while to pull dependencies if you haven't had them already.
This builds the plugin in `./build/`

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Load the plugin and create a task, see example in [Examples](#examples).

#### Configuration parameters
It's possible to provide configuration to plugin via task manifest.

In order to setup Docker Remote API endpoint and procfs path in **workflow** section of a task configuration file it is necessary to include following:

    workflow: 
      collect: 
        config: 
          /intel/docker: 
            endpoint: "<DOCKER_REMOTE_API_ENDPOINT>"
            procfs: "<PATH_TO_PROCFS>"

where *DOCKER_REMOTE_API_ENDPOINT* is an endpoint that is being used to communicate with Docker daemon via Docker Remote API,
where *PATH_TO_PROCFS* is a path to proc filesystem on host.

For more information see [Docker Remote API reference](https://docs.docker.com/engine/reference/api/docker_remote_api/)

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
$ snapteld -l 1 -t 0
```

In another terminal window download and load plugins:
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-docker/latest/linux/x86_64/snap-plugin-collector-docker
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ chmod 755 snap-plugin-*
$ snaptel plugin load snap-plugin-collector-docker
$ snaptel plugin load snap-plugin-publisher-file
```

You can list all of available metrics:
```
$ snaptel metric list
```

Download an [example task file](examples/tasks/docker-file.json) and load it:
```
$ curl -sfLO https://raw.githubusercontent.com/intelsdi-x/snap-plugin-collector-docker/master/examples/tasks/docker-file.json
$ snaptel task create -t docker-file.json
Using task manifest to create task
Task created
ID: 02dd7ff4-8106-47e9-8b86-70067cd0a850
Name: Task-02dd7ff4-8106-47e9-8b86-70067cd0a850
State: Running
```

See  output from snaptel task watch <task_id>

(notice, that below only the fragment of task watcher output has been presented)

```
$ snaptel task watch 02dd7ff4-8106-47e9-8b86-70067cd0a850
Watching Task (02dd7ff4-8106-47e9-8b86-70067cd0a850):
NAMESPACE                                                                    DATA      		TIMESTAMP
/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/total_usage           2.146646e+07       2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/kernel_mode           1e+07              2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/7720efd76bb8/cgroups/cpu_stats/cpu_usage/user_mode             0                  2016-06-21 12:44:09.552146203 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/total_usage           2.146646e+07       2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/kernel_mode           1e+07              2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/ad5221e8ae73/cgroups/cpu_stats/cpu_usage/user_mode             0                  2016-06-21 12:44:09.552146203 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/total_usage                   2.88984998661e+12  2016-06-21 12:44:09.551811277 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/kernel_mode                   6.38e+11            2016-06-21 12:44:09.552107446 +0200 CEST
/intel/docker/root/cgroups/cpu_stats/cpu_usage/user_mode                     9.4397e+11          2016-06-21 12:44:09.552146203 +0200 CEST
```
(Keys `ctrl+c` terminate task watcher)

These data are published to file and stored there (in this example in `/tmp/snap-docker-file.log`).

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release.

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-docker/issues) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-docker/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap.

To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [Slack](http://slack.snap-telemetry.io).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author:       [Marcin Krolik](https://github.com/marcin-krolik)
* Co-authors:   [Izabella Raulin](https://github.com/IzabellaRaulin), [Marcin Olszewski](https://github.com/marcintao)

**Thank you!** Your contribution is incredibly important to us.
