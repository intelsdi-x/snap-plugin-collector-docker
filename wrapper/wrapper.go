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
*/

package wrapper

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
)

// Cgroups2Stats holds pointer to appropriate cgroup type (defined in lib `opencontainers`) under cgroup name as a key
var Cgroups2Stats = map[string]Stats{
	"cpuset":  &CpuSet{},
	"shares":  &Shares{},
	"cpu":     &fs.CpuGroup{},
	"cpuacct": &fs.CpuacctGroup{},
	"memory":  &fs.MemoryGroup{},
	"devices": &fs.DevicesGroup{},
}

// Stats provides method to get cgroups statistics
type Stats interface {
	GetStats(path string, stats *cgroups.Stats) error
}

// Specification holds docker container specification
type Specification struct {
	Status     string            `json:"status"`
	Created    string            `json:"creation_time"`
	Image      string            `json:"image_name"`
	SizeRw     int64             `json:"size_rw"`
	SizeRootFs int64             `json:"size_root_fs"`
	Labels     map[string]string `json:"labels"`
}

// Statistics holds all available statistics: network, tcp/tcp6, cgroups and filesystem
type Statistics struct {
	Network         []NetworkInterface             `json:"network"`
	Connection      TcpInterface                   `json:"connection"` //TCP, TCP6 connection stats
	CgroupStats     *cgroups.Stats                 `json:"cgroups"`
	CgroupsExtended *CgroupExtended                `json:"cgroups_extended"`
	Filesystem      map[string]FilesystemInterface `json:"filesystem"`
}

// CgroupsExtended holds additional group statistics like cpuset and shares
// These stats are not supported by libcontainers lib
type CgroupExtended struct {
	CpuSet CpuSet `json:"cpuset"`
	Shares Shares `json:"shares"`
}

// GetExtendedStats extracts CPUs and memory nodes a group can access
func (cs *CpuSet) GetExtendedStats(path string, ext *CgroupExtended) error {
	cpus, err := ioutil.ReadFile(filepath.Join(path, "cpuset.cpus"))
	if err != nil {
		return err
	}
	mems, err := ioutil.ReadFile(filepath.Join(path, "cpuset.mems"))
	if err != nil {
		return err
	}
	memmig, err := ioutil.ReadFile(filepath.Join(path, "cpuset.memory_migrate"))
	if err != nil {
		return err
	}
	cpuexc, err := ioutil.ReadFile(filepath.Join(path, "cpuset.cpu_exclusive"))
	if err != nil {
		return err
	}
	memexc, err := ioutil.ReadFile(filepath.Join(path, "cpuset.mem_exclusive"))
	if err != nil {
		return err
	}

	ext.CpuSet.Cpus = string(cpus)
	ext.CpuSet.Mems = string(mems)

	ext.CpuSet.MemoryMigrate, err = strconv.ParseUint(strings.Trim(string(memmig), "\n"), 10, 64)
	if err != nil {
		return err
	}

	ext.CpuSet.CpuExclusive, err = strconv.ParseUint(strings.Trim(string(cpuexc), "\n"), 10, 64)
	if err != nil {
		return err
	}

	ext.CpuSet.MemoryExclusive, err = strconv.ParseUint(strings.Trim(string(memexc), "\n"), 10, 64)
	if err != nil {
		return err
	}

	return nil
}

// GetStats is not used. It is defined only to implement interface for CpuSet as a member of Cgroups2Stats
// It is temporary solution until openlibcontainers lib provides mechanism to extract cpuset values from cgroups
func (cs *CpuSet) GetStats(path string, stats *cgroups.Stats) error {
	return nil
}

// CpuSet stores information regarding subsystem assignment of individual CPUs and memory nodes
type CpuSet struct {
	Cpus            string `json:"cpus"`
	Mems            string `json:"mems"`
	MemoryMigrate   uint64 `json:"memory_migrate"`
	CpuExclusive    uint64 `json:"cpu_exclusive"`
	MemoryExclusive uint64 `json:"memory_exclusive"`
}

// Shares stores information regarding relative share of CPU time available to the tasks in a group
type Shares struct {
	Cpu uint64 `json:"cpu"`
}

// GetExtendedStats extracts integer value from cpu.shares file in given group
func (s *Shares) GetExtendedStats(path string, ext *CgroupExtended) error {
	cpu, err := ioutil.ReadFile(filepath.Join(path, "cpu.shares"))
	if err != nil {
		return err
	}
	ext.Shares.Cpu, err = strconv.ParseUint(strings.Trim(string(cpu), "\n"), 10, 64)
	if err != nil {
		return err
	}

	return nil
}

// GetStats is not used. It is defined only to implement interface for Shares as a member of Cgroups2Stats
// It is temporary solution until openlibcontainers lib provides mechanism to extract shares from cgroups
func (s *Shares) GetStats(path string, stats *cgroups.Stats) error {
	return nil
}

// NetworkInterface holds name of network interface and its statistics (rx_bytes, tx_bytes, etc.)
type NetworkInterface struct {
	// Name is the name of the network interface.
	Name string `json:"-"`

	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

// FilesystemInterface holds statistics about filesystem device, capacity, usage, etc.
type FilesystemInterface struct {
	// The block device name associated with the filesystem
	Device string `json:"device_name"`

	// Type of the filesystem
	Type string `json:"type"`

	// Number of bytes that can be consumed on this filesystem
	Limit uint64 `json:"capacity"`

	// Number of bytes that is consumed on this filesystem
	Usage uint64 `json:"usage"`

	// Base Usage that is consumed by the container's writable layer
	BaseUsage uint64 `json:"base_usage"`

	// Number of bytes available for non-root user
	Available uint64 `json:"available"`

	// Number of available Inodes
	InodesFree uint64 `json:"inodes_free"`

	// This is the total number of reads completed successfully
	ReadsCompleted uint64 `json:"reads_completed"`

	// This is the total number of reads merged successfully. This field lets you know how often this was done
	ReadsMerged uint64 `json:"reads_merged"`

	// This is the total number of sectors read successfully
	SectorsRead uint64 `json:"sectors_read"`

	// This is the total number of milliseconds spent reading
	ReadTime uint64 `json:"read_time"`

	// This is the total number of writes completed successfully
	WritesCompleted uint64 `json:"writes_completed"`

	// This is the total number of writes merged successfully
	WritesMerged uint64 `json:"writes_merged"`

	// This is the total number of sectors written successfully
	SectorsWritten uint64 `json:"sectors_written"`

	// This is the total number of milliseconds spent writing
	WriteTime uint64 `json:"write_time"`

	// Number of I/Os currently in progress
	IoInProgress uint64 `json:"io_in_progress"`

	// Number of milliseconds spent doing I/Os
	IoTime uint64 `json:"io_time"`

	// weighted number of milliseconds spent doing I/Os
	// This field is incremented at each I/O start, I/O completion, I/O
	// merge, or read of these stats by the number of I/Os in progress
	// (field 9) times the number of milliseconds spent doing I/O since the
	// last update of this field.  This can provide an easy measure of both
	// I/O completion time and the backlog that may be accumulating.
	WeightedIoTime uint64 `json:"weighted_io_time"`
}

// TcpInterface holds tcp and tcp6 statistics
type TcpInterface struct {
	Tcp  TcpStat `json:"tcp"`  // TCP connection stats (Established, Listen, etc.)
	Tcp6 TcpStat `json:"tcp6"` // TCP6 connection stats (Established, Listen, etc.)
}

// TcpStat holds statistics about count of connections in different states
type TcpStat struct {
	//Count of TCP connections in state "Established"
	Established uint64 `json:"established"`
	//Count of TCP connections in state "Syn_Sent"
	SynSent uint64 `json:"syn_sent"`
	//Count of TCP connections in state "Syn_Recv"
	SynRecv uint64 `json:"syn_recv"`
	//Count of TCP connections in state "Fin_Wait1"
	FinWait1 uint64 `json:"fin_wait1"`
	//Count of TCP connections in state "Fin_Wait2"
	FinWait2 uint64 `json:"fin_wait2"`
	//Count of TCP connections in state "Time_Wait
	TimeWait uint64 `json:"time_wait"`
	//Count of TCP connections in state "Close"
	Close uint64 `json:"close"`
	//Count of TCP connections in state "Close_Wait"
	CloseWait uint64 `json:"close_wait"`
	//Count of TCP connections in state "Listen_Ack"
	LastAck uint64 `json:"last_ack"`
	//Count of TCP connections in state "Listen"
	Listen uint64 `json:"listen"`
	//Count of TCP connections in state "Closing"
	Closing uint64 `json:"closing"`
}

var listOfMemoryStats = []string{
	"active_anon", "active_file", "inactive_anon", "inactive_file", "cache", "dirty", "swap",
	"hierarchical_memory_limit", "hierarchical_memsw_limit", "mapped_file", "pgfault", "pgmajfault", "pgpgin",
	"pgpgout", "rss", "rss_huge", "total_active_anon", "total_active_file", "total_cache", "total_dirty",
	"total_inactive_anon", "total_inactive_file", "total_mapped_file", "total_pgfault", "total_pgmajfault",
	"total_pgpgin", "total_pgpgout", "total_rss", "total_rss_huge", "total_swap", "total_unevictable",
	"total_writeback", "unevictable", "working_set", "writeback",
}

// NewStatistics returns pointer to initialized Statistics
func NewStatistics() *Statistics {
	return &Statistics{
		Network:         []NetworkInterface{},
		CgroupStats:     newCgroupsStats(),
		CgroupsExtended: &CgroupExtended{},
		Connection: TcpInterface{
			Tcp:  TcpStat{},
			Tcp6: TcpStat{},
		},
		Filesystem: map[string]FilesystemInterface{},
	}
}

func newCgroupsStats() *cgroups.Stats {
	cgroupStats := cgroups.NewStats()
	// set names of default memory statistics which are supported
	for _, memstatName := range listOfMemoryStats {
		cgroupStats.MemoryStats.Stats[memstatName] = 0
	}
	return cgroupStats
}
