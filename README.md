##Pulse Docker Collector plugin
<!---
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

#Description
Plugin collects Docker runtime  metrics using remote API and libcontainers cgroup stats.

#Assumptions
* Linux kernel version >= 2.6.12-rc2
* Docker installed

#Metrics
 - /intel
	- /linux
		- /docker
			- /<docker_id>
				- /cpu_stats
					- /cpu_usage
						- /total_usage
						- /usage_in_kernelmode
						- /usage_in_usermode
						- /percpu_usage
							/0
							/1
							..
							/N
					- throttling_data
						- /periods
						- /throttled_periods
						- /throttled_time
				- /memory_stats
					- /cache
					- /usage
						- /usage
						- /max_usage
						- /failcnt
					- /swap_usage
						- /usage
						- /max_usage
						- /failcnt
					- /kernel_usage
						- /usage
						- /max_usage
						- /failcnt
					- /stats
						- /cache
						- /rss
						- /rss_huge
						- /mapped_file
						- /writeback	
						- /pgpgin
						- /pgpgout
						- /pgfault
						- /pgmajfault
						- /inactive_anon
						- /active_anon
						- /inactive_file
						- /active_file
						- /unevictable
						- /hierarchical_memory_limit
						- /total_cache
						- /total_rss
						- /total_rss_huge
						- /total_mapped_file
						- /total_writeback
						- /total_pgpgin
						- /total_pgpgout
						- /total_pgfault
						- /total_pgmajfault
						- /total_inactive_anon
						- /total_active_anon
						- /total_inactive_file
						- /total_active_file
						- /total_unevictable
				- /blkio_stats
					- /io_service_bytes_recursive
					- /io_serviced_recursive
					- /io_queue_recursive
					- /io_service_time_recursive
					- /io_wait_time_recursive
					- /io_merged_recursive
					- /io_time_recursive
					- /sectors_recursive
				- /hugetlb_stats
					- /2MB
						- /usage
						- /max_usage
						- /failcnt

