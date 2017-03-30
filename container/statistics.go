// +build linux

/*
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

This file incorporates work covered by the following copyright and permission notice:
	Copyright 2014 Docker, Inc.
Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package contains code from OCI/opencontainers (https://github.com/opencontainers/runc) with following:
// - structure Statistics and its compositions

package container

import "fmt"

type GetStatOpt map[string]interface{}

func (opt GetStatOpt) GetStringValue(key string) (string, error) {
	val, exists := opt[key]
	if exists {
		if res, ok := val.(string); ok {
			return res, nil
		}
		return "", fmt.Errorf("value %v seems not be of string type", val)

	}

	return "", fmt.Errorf("could not find value for key %s", key)
}

func (opt GetStatOpt) GetIntValue(key string) (int, error) {
	val, exists := opt[key]
	if exists {
		if res, ok := val.(int); ok {
			return res, nil
		}
		return 0, fmt.Errorf("value %v seems not be of integer type", val)
	}

	return 0, fmt.Errorf("could not find value for key %s", key)
}

func (opt GetStatOpt) GetBoolValue(key string) (bool, error) {
	val, exists := opt[key]
	if exists {
		return val.(bool), nil
	}

	return false, fmt.Errorf("could not find value for key %s", key)
}

type StatGetter interface {
	GetStats(*Statistics, GetStatOpt) error
}

type ContainerData struct {
	ID string `json:"-"`
	// Basic info about the container (status, creation time, image name, etc.)
	Specification Specification `json:"spec,omitempty"`

	// Container's statistics (cpu usage, memory usage, network stats, etc.)
	Stats *Statistics `json:"stats,omitempty"`
}

type Statistics struct {
	Cgroups    *Cgroups                       `json:"cgroups,omitempty"`
	Network    []NetworkInterface             `json:"network,omitempty"`
	Connection TcpInterface                   `json:"connection,omitempty"`
	Filesystem map[string]FilesystemInterface `json:"filesystem,omitempty"`
}

// Specification holds docker container specification
type Specification struct {
	Status     string            `json:"status,omitempty"`
	Created    string            `json:"creation_time,omitempty"`
	Image      string            `json:"image_name,omitempty"`
	SizeRw     int64             `json:"size_rw,omitempty"`
	SizeRootFs int64             `json:"size_root_fs,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

type Cgroups struct {
	CpuStats     CpuStats                `json:"cpu_stats,omitempty"`
	MemoryStats  MemoryStats             `json:"memory_stats,omitempty"`
	BlkioStats   BlkioStats              `json:"blkio_stats, omitempty"`
	HugetlbStats map[string]HugetlbStats `json:"hugetlb_stats,omitempty"`
	PidsStats    PidsStats               `json:"pids_stats,omitempty"`
	CpuSetStats  CpuSetStats             `json:"cpuset_stats,omitempty"`
}

type CpuStats struct {
	CpuUsage       CpuUsage       `json:"cpu_usage,omitempty"`
	ThrottlingData ThrottlingData `json:"throttling_data,omitempty"`
	CpuShares      uint64         `json:"cpu_shares,omitempty"`
}

type CpuUsage struct {
	Total      uint64   `json:"total,omitempty"`
	UserMode   uint64   `json:"user_mode,omitempty"`
	KernelMode uint64   `json:"kernel_mode,omitempty"`
	PerCpu     []uint64 `json:"per_cpu,omitempty"`
}

type ThrottlingData struct {
	NrPeriods     uint64 `json:"nr_periods,omitempty"`
	NrThrottled   uint64 `json:"nr_throttled,omitempty"`
	ThrottledTime uint64 `json:"throttled_time,omitempty"`
}

type MemoryStats struct {
	Cache       uint64            `json:"cache,omitempty"`
	Usage       MemoryData        `json:"usage,omitempty"`
	SwapUsage   MemoryData        `json:"swap_usage,omitempty"`
	KernelUsage MemoryData        `json:"kernel_usage,omitempty"`
	Stats       map[string]uint64 `json:"statistics,omitempty"`
}

type MemoryData struct {
	Usage    uint64 `json:"usage,omitempty"`
	MaxUsage uint64 `json:"max_usage,omitempty"`
	Failcnt  uint64 `json:"failcnt,omitempty"`
}

type BlkioStats struct {
	// number of bytes tranferred to and from the block device
	IoServiceBytesRecursive []BlkioStatEntry `json:"io_service_bytes_recursive,omitempty"`
	IoServicedRecursive     []BlkioStatEntry `json:"io_serviced_recursive,omitempty"`
	IoQueuedRecursive       []BlkioStatEntry `json:"io_queue_recursive,omitempty"`
	IoServiceTimeRecursive  []BlkioStatEntry `json:"io_service_time_recursive,omitempty"`
	IoWaitTimeRecursive     []BlkioStatEntry `json:"io_wait_time_recursive,omitempty"`
	IoMergedRecursive       []BlkioStatEntry `json:"io_merged_recursive,omitempty"`
	IoTimeRecursive         []BlkioStatEntry `json:"io_time_recursive,omitempty"`
	SectorsRecursive        []BlkioStatEntry `json:"sectors_recursive,omitempty"`
}

type BlkioStatEntry struct {
	Major uint64 `json:"major,omitempty"`
	Minor uint64 `json:"minor,omitempty"`
	Op    string `json:"op,omitempty"`
	Value uint64 `json:"value,omitempty"`
}

type HugetlbStats struct {
	// current res_counter usage for hugetlb
	Usage uint64 `json:"usage,omitempty"`
	// maximum usage ever recorded.
	MaxUsage uint64 `json:"max_usage,omitempty"`
	// number of times htgetlb usage allocation failure.
	Failcnt uint64 `json:"failcnt,omitempty"`
}

type PidsStats struct {
	Current uint64 `json:"current,omitempty"`
	Limit   uint64 `json:"limit,omitempty"`
}

// CpuSet stores information regarding subsystem assignment of individual CPUs and memory nodes
type CpuSetStats struct {
	Cpus            string `json:"cpus,omitempty"`
	Mems            string `json:"mems,omitempty"`
	MemoryMigrate   uint64 `json:"memory_migrate,omitempty"`
	CpuExclusive    uint64 `json:"cpu_exclusive,omitempty"`
	MemoryExclusive uint64 `json:"memory_exclusive,omitempty"`
}

// NetworkInterface holds name of network interface and its statistics (rx_bytes, tx_bytes, etc.)
type NetworkInterface struct {
	// Name is the name of the network interface.
	Name string `json:"-"`

	RxBytes   uint64 `json:"rx_bytes,omitempty"`
	RxPackets uint64 `json:"rx_packets,omitempty"`
	RxErrors  uint64 `json:"rx_errors,omitempty"`
	RxDropped uint64 `json:"rx_dropped,omitempty"`
	TxBytes   uint64 `json:"tx_bytes,omitempty"`
	TxPackets uint64 `json:"tx_packets,omitempty"`
	TxErrors  uint64 `json:"tx_errors,omitempty"`
	TxDropped uint64 `json:"tx_dropped,omitempty"`
}

// FilesystemInterface holds statistics about filesystem device, capacity, usage, etc.
type FilesystemInterface struct {
	// The block device name associated with the filesystem
	Device string `json:"device_name,omitempty"`

	// Type of the filesystem
	Type string `json:"type,omitempty"`

	// Number of bytes that can be consumed on this filesystem
	Limit uint64 `json:"capacity,omitempty"`

	// Number of bytes that is consumed on this filesystem
	Usage uint64 `json:"usage,omitempty"`

	// Base Usage that is consumed by the container's writable layer
	BaseUsage uint64 `json:"base_usage,omitempty"`

	// Number of bytes available for non-root user
	Available uint64 `json:"available,omitempty"`

	// Number of available Inodes
	InodesFree uint64 `json:"inodes_free,omitempty"`

	// This is the total number of reads completed successfully
	ReadsCompleted uint64 `json:"reads_completed,omitempty"`

	// This is the total number of reads merged successfully. This field lets you know how often this was done
	ReadsMerged uint64 `json:"reads_merged,omitempty"`

	// This is the total number of sectors read successfully
	SectorsRead uint64 `json:"sectors_read,omitempty"`

	// This is the total number of milliseconds spent reading
	ReadTime uint64 `json:"read_time,omitempty"`

	// This is the total number of writes completed successfully
	WritesCompleted uint64 `json:"writes_completed,omitempty"`

	// This is the total number of writes merged successfully
	WritesMerged uint64 `json:"writes_merged,omitempty"`

	// This is the total number of sectors written successfully
	SectorsWritten uint64 `json:"sectors_written,omitempty"`

	// This is the total number of milliseconds spent writing
	WriteTime uint64 `json:"write_time,omitempty"`

	// Number of I/Os currently in progress
	IoInProgress uint64 `json:"io_in_progress,omitempty"`

	// Number of milliseconds spent doing I/Os
	IoTime uint64 `json:"io_time,omitempty"`

	// weighted number of milliseconds spent doing I/Os
	// This field is incremented at each I/O start, I/O completion, I/O
	// merge, or read of these stats by the number of I/Os in progress
	// (field 9) times the number of milliseconds spent doing I/O since the
	// last update of this field.  This can provide an easy measure of both
	// I/O completion time and the backlog that may be accumulating.
	WeightedIoTime uint64 `json:"weighted_io_time,omitempty"`
}

// TcpInterface holds tcp and tcp6 statistics
type TcpInterface struct {
	Tcp  TcpStat `json:"tcp,omitempty"`  // TCP connection stats (Established, Listen, etc.)
	Tcp6 TcpStat `json:"tcp6,omitempty"` // TCP6 connection stats (Established, Listen, etc.)
}

// TcpStat holds statistics about count of connections in different states
type TcpStat struct {
	//Count of TCP connections in state "Established"
	Established uint64 `json:"established,omitempty"`
	//Count of TCP connections in state "Syn_Sent"
	SynSent uint64 `json:"syn_sent,omitempty"`
	//Count of TCP connections in state "Syn_Recv"
	SynRecv uint64 `json:"syn_recv,omitempty"`
	//Count of TCP connections in state "Fin_Wait1"
	FinWait1 uint64 `json:"fin_wait1,omitempty"`
	//Count of TCP connections in state "Fin_Wait2"
	FinWait2 uint64 `json:"fin_wait2,omitempty"`
	//Count of TCP connections in state "Time_Wait
	TimeWait uint64 `json:"time_wait,omitempty"`
	//Count of TCP connections in state "Close"
	Close uint64 `json:"close,omitempty"`
	//Count of TCP connections in state "Close_Wait"
	CloseWait uint64 `json:"close_wait,omitempty"`
	//Count of TCP connections in state "Listen_Ack"
	LastAck uint64 `json:"last_ack,omitempty"`
	//Count of TCP connections in state "Listen"
	Listen uint64 `json:"listen,omitempty"`
	//Count of TCP connections in state "Closing"
	Closing uint64 `json:"closing,omitempty"`
}

// NewStatistics returns pointer to initialized Statistics
func NewStatistics() *Statistics {
	return &Statistics{
		Network: []NetworkInterface{},
		Cgroups: newCgroupsStats(),
		Connection: TcpInterface{
			Tcp:  TcpStat{},
			Tcp6: TcpStat{},
		},
		Filesystem: map[string]FilesystemInterface{},
	}
}

func newCgroupsStats() *Cgroups {
	cgroups := Cgroups{
		MemoryStats:  MemoryStats{Stats: make(map[string]uint64)},
		HugetlbStats: make(map[string]HugetlbStats),
	}
	for _, memstatName := range listOfMemoryStats {
		cgroups.MemoryStats.Stats[memstatName] = 0
	}
	return &cgroups
}

var listOfMemoryStats = []string{
	"active_anon", "active_file", "inactive_anon", "inactive_file", "cache", "dirty", "swap",
	"hierarchical_memory_limit", "hierarchical_memsw_limit", "mapped_file", "pgfault", "pgmajfault", "pgpgin",
	"pgpgout", "rss", "rss_huge", "total_active_anon", "total_active_file", "total_cache", "total_dirty",
	"total_inactive_anon", "total_inactive_file", "total_mapped_file", "total_pgfault", "total_pgmajfault",
	"total_pgpgin", "total_pgpgout", "total_rss", "total_rss_huge", "total_swap", "total_unevictable",
	"total_writeback", "unevictable", "working_set", "writeback",
}
