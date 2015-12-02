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

[![Build Status](https://travis-ci.com/intelsdi-x/snap-plugin-collector-docker.svg?token=HoxHq3yqBGpySzRd5XUm&branch=master)](https://travis-ci.com/intelsdi-x/snap-plugin-collector-docker)

# snap Docker Collector Plugin

Plugin collects runtime metrics from docker containers on host machine. It gathers information about resource usage and perfromance characteristics of running containers. 

It's used in the [snap framework](http://github.com/intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license-and-authors)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need docker engine installed. Visit [Install Docker Engine](https://docs.docker.com/engine/installation/) for detailed instructions how to do it.

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
This builds the plugin in `/build/rootfs/`

### Configuration and Usage
* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported  
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`

All metrics gathered by this plugin are exposed by [cgroups](https://www.kernel.org/doc/Documentation/cgroups/cgroups.txt) 

## Documentation

### Collected Metrics
This plugin has the ability to gather the following metrics:

Namespace | Data Type | Description (optional)
----------|-----------|-----------------------
/intel/linux/docker/cpu_stats/cpu_usage/total_usage | uint64 | Total CPU time consumed
/intel/linux/docker/cpu_stats/cpu_usage/usage_in_kernelmode | uint64 | CPU time consumed by tasks in system (kernel) mode
/intel/linux/docker/cpu_stats/cpu_usage/usage_in_usermode | uint64 | CPU time consumed by tasks in user mode
/intel/linux/docker/cpu_stats/cpu_usage/percpu_usage/\<N\> | uint64 | CPU time consumed on each N-th CPU by all tasks
/intel/linux/docker/cpu_stats/throttling_data/periods | uint64 | number of period intervals that have elapsed
/intel/linux/docker/cpu_stats/throttling_data/throttled_periods | uint64 | number of times tasks in a cgroup have been throttled
/intel/linux/docker/cpu_stats/throttling_data/throttled_time | uint64 | total time duration for which tasks in a cgroup have been throttled
/intel/linux/docker/memory_stats/cache | uint64 | page cache including tmpfs
/intel/linux/docker/memory_stats/usage/usage | uint64 | reports the total current memory usage by processes in the cgroup
/intel/linux/docker/memory_stats/usage/max_usage | uint64 | reports the maximum memory used by processes in the cgroup
/intel/linux/docker/memory_stats/usage/failcnt | uint64 | reports the number of times that the memory limit has reached the value set in memory.limit_in_bytes
/intel/linux/docker/memory_stats/swap_usage/usage | uint64 | reports the total swap space usage by processes in the cgroup
/intel/linux/docker/memory_stats/swap_usage/max_usage | uint64 | reports the maximum swap space used by processes in the cgroup
/intel/linux/docker/memory_stats/swap_usage/failcnt | uint64 | reports the number of times the swap space limit has reached the value set in memorysw.limit_in_bytes
/intel/linux/docker/memory_stats/kernel_usage/usage | uint64 | reports the total kernel memory allocation by processes in the cgroup
/intel/linux/docker/memory_stats/kernel_usage/max_usage | uint64 | reports the maximum kernel memory allocation by processes in the cgroup
/intel/linux/docker/memory_stats/kernel_usage/failcnt | uint64 | reports the number of times the kernel memory allocation has reached the value set in kmem.limit_in_bytes
/intel/linux/docker/memory_stats/stats/cache | uint64 | number of bytes of page cache memory
/intel/linux/docker/memory_stats/stats/rss | uint64 | number of bytes of anonymous and swap cache memory
/intel/linux/docker/memory_stats/stats/rss_huge | uint64 | number of bytes of anonymous transparent hugepages
/intel/linux/docker/memory_stats/stats/mapped_file | uint64 | number of bytes of mapped file (includes tmpfs/shmem)
/intel/linux/docker/memory_stats/stats/writeback | uint64 | number of bytes of file/anon cache that are queued for syncing to disk
/intel/linux/docker/memory_stats/stats/pgpgin | uint64 | number of charging events to the memory cgroup
/intel/linux/docker/memory_stats/stats/pgpgout | uint64 | number of uncharging events to the memory cgroup
/intel/linux/docker/memory_stats/stats/pgfault | uint64 | number of page faults which happened since the creation of the cgroup
/intel/linux/docker/memory_stats/stats/pgmajfault | uint64 | number of page major faults which happened since the creation of the cgroup
/intel/linux/docker/memory_stats/stats/active_anon | uint64 | number of bytes of anonymous and swap cache memory on active LRU list
/intel/linux/docker/memory_stats/stats/inactive_anon | uint64 | number of bytes of anonymous and swap cache memory on inactive LRU list
/intel/linux/docker/memory_stats/stats/active_file | uint64 | number of bytes of file-backed memory on active LRU list
/intel/linux/docker/memory_stats/stats/inactive_file | uint64 | number of bytes of file-backed memory on inactive LRU list
/intel/linux/docker/memory_stats/stats/unevictable | uint64 | number of bytes of memory that cannot be reclaimed
/intel/linux/docker/memory_stats/stats/hierarchical_memory_limit | uint64 | of bytes of memory limit with regard to hierarchy under which the memory cgroup is
/intel/linux/docker/memory_stats/stats/total_\<counter\> | uint64 | hierarchical version of \<counter\>, which in addition to the cgroup's own value includes the sum of all hierarchical children's values of \<counter\>
/intel/linux/docker/blkio_stats/io_service_bytes_recursive | uint64 | number of bytes transferred to/from the disk from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_service_recursive | uint64 | number of IOs (bio) issued to the disk from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_queue_recursive | uint64 | number of  requests queued up at any given instant from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_service_time_recursive | uint64 | amount of time between request dispatch and request completion from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_wait_time_recursive | uint64 | amount of time the IOs for this cgroup spent waiting in the scheduler queues for service from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_merged_recursive | uint64 | number of bios/requests merged into requests belonging to all the descendant cgroups
/intel/linux/docker/blkio_stats/io_time_recursive | uint64 | disk time allocated to all devices from all the descendant cgroups
/intel/linux/docker/blkio_stats/io_sectors_recursive | uint64 | number of sectors transferred to/from disk bys from all the descendant cgroups
/intel/linux/docker/hugetlb_stats/\<hugepagesize\>/usage | uint64 | show current usage for "hugepagesize" hugetlb
/intel/linux/docker/hugetlb_stats/\<hugepagesize\>/max_usage | uint64 | show max "hugepagesize" hugetlb usage recorded
/intel/linux/docker/hugetlb_stats/\<hugepagesize\>/failcnt | uint64 | show the number of allocation failure due to HugeTLB limit

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-docker/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-docker/pulls).

## Community Support
This repository is one of **many** plugins in the **snap Framework**: a powerful telemetry agent framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Marcin Krolik](https://github.com/marcin-krolik)

**Thank you!** Your contribution is incredibly important to us.
