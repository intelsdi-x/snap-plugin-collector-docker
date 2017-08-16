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

# snap plugin collector - Docker

## Collected Metrics
This plugin has the ability to gather the following metrics:

</br>
a) **docker container specification**

The prefix of metric's namespace is `/intel/docker/<docker_id>/spec/`

(e.g. /intel/docker/12345/spec/creation_time)

Namespace | Data Type | Description
----------|-----------|-----------------------
creation_time | string | The time when the container was started
image_name | string | The name of docker image that container has been created of
size_root_fs | uint64 | The total size of all the files in the container, in bytes
size_rw | uint64 | The size of the files which have been created or changed in reference to the container base image. After container creation, this should be zero and will increase as files are created/modified
status | string | The status of docker container 
labels/\<label_key\>/value | string | The value of the container's label under the key
</br>

Notice, that these kind of metrics are not available for host of docker containers.

b) **cgroups statistics**

The prefix of metric's namespace is `/intel/docker/<docker_id_or_root>/stats/cgroups/`

(e.g. /intel/docker/12345/stats/cgroups/memory_stats/cache)


Namespace | Data Type | Description
----------|-----------|-----------------------
cpu_stats/cpu_usage/total | uint64 | The total CPU time consumed
cpu_stats/cpu_usage/kernel_mode | uint64 | CPU time consumed by tasks in system (kernel) mode
cpu_stats/cpu_usage/user_mode | uint64 | CPU time consumed by tasks in user mode
cpu_stats/cpu_usage/per_cpu/\<N\>/value | uint64 | CPU time consumed on each N-th CPU by all tasks
cpu_stats/throttling_data/nr_periods | uint64 | The number of period intervals that have elapsed
cpu_stats/throttling_data/nr_throttled | uint64 | The number of times tasks in a cgroup have been throttled
cpu_stats/throttling_data/throttled_time | uint64 | The total time duration for which tasks in a cgroup have been throttled
cpu_stats/cpu_shares | uint64 | The relative share of CPU time available to the tasks in a cgroup
cpuset_stats/cpu_exclusive | uint64 | Flag (0 or 1) that specifies whether cpusets other than this one and its parents and children can share the CPUs specified for this cpuset
cpuset_stats/memory_exclusive | uint64 | Flag (0 or 1) that specifies whether other cpusets can share the memory nodes specified for the cpuset
cpuset_stats/memory_migrate | uint64 | Flag (0 or 1) that specifies whether a page in memory should migrate to a new node if the values in cpuset.mems change
cpuset_stats/cpus | string | CPUs numbers that tasks in this cgroup are permitted to access
cpuset_stats/mems | string | Memory nodes that tasks in this cgroup are permitted to access
| |
pids_stats/current | uint64 | The current number of PID in the cgroup
pids_stats/limit | uint64 | The maximum number of PIDs in the cgroup
| |
memory_stats/cache | uint64 | Page cache including tmpfs
memory_stats/usage/usage | uint64 | Total current memory usage by processes in the cgroup
memory_stats/usage/max_usage | uint64 | The maximum memory used by processes in the cgroup
memory_stats/usage/failcnt | uint64 | The number of times that the memory limit has reached the value set in memory.limit_in_bytes
memory_stats/swap_usage/usage | uint64 | The total swap space usage by processes in the cgroup
memory_stats/swap_usage/max_usage | uint64 | The maximum swap space used by processes in the cgroup
memory_stats/swap_usage/failcnt | uint64 | The number of times the swap space limit has reached the value set in memorysw.limit_in_bytes
memory_stats/kernel_usage/usage | uint64 | The total kernel memory allocation by processes in the cgroup
memory_stats/kernel_usage/max_usage | uint64 | The maximum kernel memory allocation by processes in the cgroup
memory_stats/kernel_usage/failcnt | uint64 | The number of times the kernel memory allocation has reached the value set in kmem.limit_in_bytes
memory_stats/statistics/active_anon | uint64 | The number of bytes of anonymous and swap cache memory on active LRU list
memory_stats/statistics/active_file | uint64 | The number of bytes of file-backed memory on active LRU list
memory_stats/statistics/cache | uint64 | The number of bytes of page cache memory
memory_stats/statistics/dirty | uint64 | The number of bytes that are waiting to get written back to the disk
memory_stats/statistics/hierarchical_memory_limit | uint64 | The number of bytes of memory limit with regard to hierarchy under which the memory cgroup is
memory_stats/statistics/hierarchical_memsw_limit | uint64 | The number of bytes of memory+swap limit with regard to hierarchy under which memory cgroup is
memory_stats/statistics/inactive_anon | uint64 | The number of bytes of anonymous and swap cache memory on inactive LRU list
memory_stats/statistics/inactive_file | uint64 | The number of bytes of file-backed memory on inactive LRU list
memory_stats/statistics/mapped_file | uint64 | The number of bytes of mapped file (includes tmpfs/shmem)
memory_stats/statistics/pgfault | uint64 | The number of page faults which happened since the creation of the cgroup
memory_stats/statistics/pgmajfault | uint64 | The number of page major faults which happened since the creation of the cgroup
memory_stats/statistics/pgpgin | uint64 | The number of charging events to the memory cgroup
memory_stats/statistics/pgpgout | uint64 | The number of uncharging events to the memory cgroup
memory_stats/statistics/rss | uint64 | The number of bytes of anonymous and swap cache memory
memory_stats/statistics/rss_huge | uint64 | The number of bytes of anonymous transparent hugepages
memory_stats/statistics/swap | uint64 | The amount of swap currently used by the processes in this cgroup
memory_stats/statistics/unevictable | uint64 | The number of bytes of memory that cannot be reclaimed
memory_stats/statistics/working_set | uint64 | The total number of bytes of memory that is being used and not easily dropped by the kernel
memory_stats/statistics/writeback | uint64 | The number of bytes of file/anon cache that are queued for syncing to disk
memory_stats/statistics/total_active_anon | uint64 | The total number of bytes of anonymous and swap cache memory on active LRU list <sup>(1)</sup>
memory_stats/statistics/total_active_file | uint64 | The total number of bytes of file-backed memory on active LRU list <sup>(1)</sup>
memory_stats/statistics/total_cache | uint64 | The total number of bytes of page cache memory <sup>(1)</sup>
memory_stats/statistics/total_dirty | uint64 | The total number of bytes that are waiting to get written back to the disk <sup>(1)</sup>
memory_stats/statistics/total_inactive_anon | uint64 | The total number of bytes of anonymous and swap cache memory on inactive LRU list <sup>(1)</sup>
memory_stats/statistics/total_inactive_file | uint64 | The total number of bytes of file-backed memory on inactive LRU list <sup>(1)</sup>
memory_stats/statistics/total_mapped_file | uint64 | The total number of bytes of mapped file (includes tmpfs/shmem) <sup>(1)</sup>
memory_stats/statistics/total_pgfault | uint64 | The total number of page faults which happened since the creation of the cgroup <sup>(1)</sup>
memory_stats/statistics/total_pgmajfault | uint64 | The total number of page major faults which happened since the creation of the cgroup <sup>(1)</sup>
memory_stats/statistics/total_pgpgin | uint64 | The total number of charging events to the memory cgroup <sup>(1)</sup>
memory_stats/statistics/total_pgpgout | uint64 | The total number of uncharging events to the memory cgroup <sup>(1)</sup>
memory_stats/statistics/total_rss | uint64 | The total number of bytes of anonymous and swap cache memory <sup>(1)</sup>
memory_stats/statistics/total_rss_huge | uint64 | The total number of bytes of anonymous transparent hugepages <sup>(1)</sup>
memory_stats/statistics/total_swap | uint64 | The total number of bytes of swap usage <sup>(1)</sup>
memory_stats/statistics/total_unevictable | uint64 | The total number of bytes of memory that cannot be reclaimed <sup>(1)</sup>
memory_stats/statistics/total_writeback | uint64 | The total number of bytes of file/anon cache that are queued for syncing to disk <sup>(1)</sup>
| |
hugetlb_stats/\<size\>/failcnt | uint64 | The number of allocation failure due to HugeTLB limit
hugetlb_stats/\<size\>/max_usage | uint64 | Max "hugepagesize" hugetlb  usage recorded
hugetlb_stats/\<size\>/usage | uint64 | Current usage for "hugepagesize" hugetlb
| |
blkio_stats/io_merged_recursive/\<N\>/value | uint64 | Total number of bios/requests merged into requests belonging from all the descendant cgroups  
blkio_stats/io_queue_recursive/\<N\>/value | uint64 | Total number of requests queued up at any given instant from all the descendant cgroups 
blkio_stats/io_service_bytes_recursive/\<N\>/value | uint64 | Number of bytes transferred to/from the disk from all the descendant cgroups 
blkio_stats/io_service_time_recursive/\<N\>/value | uint64 | Total amount of time between request dispatch and request completion for the IOs done from all the descendant cgroups
blkio_stats/io_serviced_recursive/\<N\>/value | uint64 | Number of IOs (bio) issued to the disk from all the descendant cgroups
blkio_stats/io_time_recursive/\<N\>/value | uint64 | Disk time allocated to cgroup per device in milliseconds from all the descendant cgroups
blkio_stats/io_wait_time_recursive/\<N\>/value | uint64 | Total amount of time the IOs for this cgroup spent waiting in the scheduler queues for service from all the descendant cgroups
blkio_stats/sectors_recursive/\<N\>/value | uint64 | Number of sectors transferred to/from disk from all descendant group
blkio_stats/\<bklio_stat_name\>/major | uint64 | Major number of device <sup>(2)</sup>
blkio_stats/\<bklio_stat_name\>/minor | uint64 | Minor number of device <sup>(2)</sup>
blkio_stats/\<bklio_stat_name\>/op | string | Operation name <sup>(2)</sup>

<sup>(1)</sup> Hierarchical version of cgroups counter which in addition to the cgroup's own value includes the sum of all hierarchical children's values

<sup>(2)</sup> Each blkio statistic additionally exposes `major`, `minor` and `op` metric    

Read more about cgroups in [Kernel documentation](https://www.kernel.org/doc/Documentation/cgroup-v1/cgroups.txt)

</br>

c) **network statistics**

The prefix of metric's namespace is `/intel/docker/<docker_id_or_root>/stats/network/<interface_name_or_total>`

(e.g. /intel/docker/12345/stats/network/eth0/rx_bytes)

Namespace | Data Type | Description
----------|-----------|-----------------------
rx_bytes | uint64 | The number of bytes received over the network
rx_packets | uint64 | The number of packets received over the network
rx_dropped | uint64 | The number of bytes dropped during receiving over the network
rx_errors | uint64 | The number of errors while receiving over the network
tx_bytes | uint64 | The number of bytes sent over the network
tx_packets | uint64 | The number of packets sent over the network
tx_dropped | uint64 | The number of bytes dropped during sending over the network
tx_errors | uint64 | The number of errors while sending over the network

</br>

d) **tcp/tcp6 statistics**

The prefix of metric's namespace is `/intel/docker/<docker_id_or_root>/stats/connection/`

(e.g. /intel/docker/12345/connection/tcp/established)

Namespace | Data Type | Description
----------|-----------|-----------------------
tcp/close | uint64 |  The number of TCP connections in state "Close"
tcp/close_wait | uint64 |  The number of TCP connections in state "Close_Wait"
tcp/closing | uint64 |  The number of TCP connections in state "Closing"
tcp/established | uint64 | The number of TCP connections in state "Established"
tcp/fin_wait1 | uint64 | The number of TCP connections in state "Fin_Wait1"
tcp/fin_wait2 | uint64 | The number of TCP connections in state "Fin_Wait2"
tcp/last_ack | uint64 |  The number of TCP connections in state "Listen_Ack"
tcp/syn_recv | uint64 | The number of TCP connections in state "Syn_Recv"
tcp/syn_sent | uint64 | The number of TCP connections in state "Syn_Sent"
tcp/time_wait | uint64 | The number of TCP connections in state "Time_Wait"
tcp6/close | uint64 |  The number of TCP6 connections in state "Close"
tcp6/close_wait | uint64 |  The number of TCP6 connections in state "Close_Wait"
tcp6/closing | uint64 |  The number of TCP6 connections in state "Closing"
tcp6/established | uint64 | The number of TCP6 connections in state "Established"
tcp6/fin_wait1 | uint64 | The number of TCP6 connections in state "Fin_Wait1"
tcp6/fin_wait2 | uint64 | The number of TCP6 connections in state "Fin_Wait2"
tcp6/last_ack | uint64 |  The number of TCP6 connections in state "Listen_Ack"
tcp6/syn_recv | uint64 | The number of TCP6 connections in state "Syn_Recv"
tcp6/syn_sent | uint64 | The number of TCP6 connections in state "Syn_Sent"
tcp6/time_wait | uint64 | The number of TCP6 connections in state "Time_Wait"

</br>

e) **filesystem statistics**

The prefix of metric's namespace is `/intel/docker/<docker_id_or_root>/stats/filesystem/<device_name>`

(e.g. /intel/docker/12345/stats/filesystem/sda-1/available)

Namespace | Data Type | Description
----------|-----------|-----------------------
available | uint64 | The number of bytes available for non-root user
base_usage | uint64 | The base usage that is consumed by the container's writable layer
capacity | uint64 |  The number of bytes that can be consumed by the container on this filesystem
device_name | string | The block device name associated with the filesystem
inodes_free | uint64 | The number of available Inodes
io_in_progress | uint64 | The number of I/Os currently in progress
io_time | uint64 | The number of milliseconds spent doing I/Os
read_time | uint64 | The total number of milliseconds spent reading
reads_completed | uint64 | The total number of reads completed successfully
reads_merged | uint64 |  The total number of reads merged successfully
sectors_read | uint64 | The total number of sectors read successfully
sectors_written | uint64 | The total number of sectors written successfully
type | string |  Type of the filesystem (e.g. vfs)
usage | uint64 |  The number of bytes that is consumed by the container on this filesystem
weighted_io_time | uint64 |  The weighted number of milliseconds spent doing I/Os
write_time | uint64 | The total number of milliseconds spent writing
writes_completed | uint64 | The total number of writes completed successfully
writes_merged | uint64 | The total number of writes merged successfully
