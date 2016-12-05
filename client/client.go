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
*/

package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsouza/go-dockerclient"
	"github.com/intelsdi-x/snap-plugin-collector-docker/config"
	"github.com/intelsdi-x/snap-plugin-collector-docker/fs"
	"github.com/intelsdi-x/snap-plugin-collector-docker/network"
	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
	"github.com/opencontainers/runc/libcontainer/cgroups"
)

const (
	endpoint         string = "unix:///var/run/docker.sock"
	dockerVersionKey string = "Version"
)

// DockerClientInterface provides methods i.a. for interaction with the docker API.
type DockerClientInterface interface {
	ListContainersAsMap() (map[string]docker.APIContainers, error)
	GetStatsFromContainer(string, bool) (*wrapper.Statistics, error)
	InspectContainer(string) (*docker.Container, error)
	FindCgroupMountpoint(string) (string, error)
}

// DockerClient holds fsouza go-dockerclient instance ready for communication with the server endpoint `unix:///var/run/docker.sock`,
// cache instance which is used to store output from docker container inspect (to avoid execute inspect request multiply times, it is called only once per container)
// and diskUsageCollector which is responsible for collecting container disk usage (based on `du -u` command) in the background
type DockerClient struct {
	cl                 *docker.Client
	inspectCache       map[string]*docker.Container
	inspectMutex       sync.Mutex
	diskUsageCollector fs.DiskUsageCollector
}

type deviceInfo struct {
	device string
	major  string
	minor  string
}

// NewDockerClient returns dockerClient instance ready for communication with the server endpoint `unix:///var/run/docker.sock`
func NewDockerClient() (*DockerClient, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot initialize docker client instance with the given server endpoint `%s`, err=%v", endpoint, err)
	}

	dc := &DockerClient{
		cl:                 client,
		inspectCache:       map[string]*docker.Container{},
		diskUsageCollector: fs.DiskUsageCollector{},
	}

	dc.diskUsageCollector.Init()
	// get version of docker engine
	version, err := dc.version()
	if err != nil {
		return nil, err
	}
	config.DockerVersion = version

	return dc, nil
}

// FindCgroupMountpoint returns cgroup mountpoint of a given subsystem
func (dc *DockerClient) FindCgroupMountpoint(subsystem string) (string, error) {
	return cgroups.FindCgroupMountpoint(subsystem)
}

// GetShortID returns short container ID (12 chars)
func GetShortID(dockerID string) (string, error) {
	if len(dockerID) < 12 {
		return "", fmt.Errorf("Docker id %v is too short (the length of id should equal at least 12)", dockerID)
	}
	return dockerID[:12], nil
}

// GetStatsFromContainer returns docker containers stats: cgroups stats (cpu usage, memory usage, etc.) and network stats (tx_bytes, rx_bytes etc.);
// notes that incoming container id has to be full-length to be able to inspect container
func (dc *DockerClient) GetStatsFromContainer(id string, collectFs bool) (*wrapper.Statistics, error) {
	var (
		err        error
		pid        int
		workingSet uint64

		container = &docker.Container{}
		groupWrap = wrapper.Cgroups2Stats // wrapper for cgroup name and interface for stats extraction
		stats     = wrapper.NewStatistics()
	)

	if !isHost(id) {
		if !isFullLengthID(id) {
			return nil, fmt.Errorf("Container id %+v is not fully-length - cannot inspect container", id)
		}
		// inspect container based only on fully-length container id.
		container, err = dc.InspectContainer(id)

		if err != nil {
			return nil, err
		}
		// take docker container PID
		pid = container.State.Pid
	}

	for cg, stat := range groupWrap {
		groupPath, err := getSubsystemPath(cg, id)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot found subsystem path for cgroup=", cg, " for container id=", container)
			continue
		}

		switch cg {
		case "cpuset":
			cpuset := wrapper.CpuSet{}
			// get cpuset group stats
			err = cpuset.GetExtendedStats(groupPath, stats.CgroupsExtended)
		case "shares":
			// get cpu.shares stats
			shares := wrapper.Shares{}
			err = shares.GetExtendedStats(groupPath, stats.CgroupsExtended)
		default:
			// get cgroup stats for given docker
			err = stat.GetStats(groupPath, stats.CgroupStats)
		}

		if err != nil {
			// just log about it
			if isHost(id) {
				fmt.Fprintln(os.Stderr, "Cannot obtain cgroups statistics for host, err=", err)
			} else {
				fmt.Fprintln(os.Stderr, "Cannot obtain cgroups statistics for container: id=", id, ", image=", container.Image, ", name=", container.Name, ", err=", err)
			}
			continue
		}
	}

	// calculate additional stats memory:working_set based on memory_stats
	if totalInactiveAnon, ok := stats.CgroupStats.MemoryStats.Stats["total_inactive_anon"]; ok {
		workingSet = stats.CgroupStats.MemoryStats.Usage.Usage
		if workingSet < totalInactiveAnon {
			workingSet = 0
		} else {
			workingSet -= totalInactiveAnon
		}

		if totalInactiveFile, ok := stats.CgroupStats.MemoryStats.Stats["total_inactive_file"]; ok {
			if workingSet < totalInactiveFile {
				workingSet = 0
			} else {
				workingSet -= totalInactiveFile
			}
		}
	}
	stats.CgroupStats.MemoryStats.Stats["working_set"] = workingSet

	if !isHost(id) {
		rootFs := "/"

		stats.Network, err = network.NetworkStatsFromProc(rootFs, pid)
		if err != nil {
			// only log error message
			fmt.Fprintf(os.Stderr, "Unable to get network stats, containerID=%+v, pid %d: %v", container.ID, pid, err)
		}

		stats.Connection.Tcp, err = network.TcpStatsFromProc(rootFs, pid)
		if err != nil {
			// only log error message
			fmt.Fprintf(os.Stderr, "Unable to get tcp stats from pid %d: %v", pid, err)
		}

		stats.Connection.Tcp6, err = network.Tcp6StatsFromProc(rootFs, pid)
		if err != nil {
			// only log error message
			fmt.Fprintf(os.Stderr, "Unable to get tcp6 stats from pid %d: %v", pid, err)
		}

	} else {
		stats.Network, err = network.NetworkStatsFromRoot()
		if err != nil {
			// only log error message
			fmt.Fprintf(os.Stderr, "Unable to get network stats, containerID=%v, %v", id, err)
		}

	}
	if collectFs {
		stats.Filesystem, err = fs.GetFsStats(container)
		if err != nil {
			// only log error message
			fmt.Fprintf(os.Stderr, "Unable to get filesystem stats for docker: %v, err=%v", id, err)
		}
	}

	return stats, nil
}

// InspectContainer returns information about the container with given ID
func (dc *DockerClient) InspectContainer(id string) (*docker.Container, error) {
	dc.inspectMutex.Lock()
	defer dc.inspectMutex.Unlock()

	// check if the inspect info is already stored in inspectCache
	if info, haveInfo := dc.inspectCache[id]; haveInfo {
		return info, nil
	}

	info, err := dc.cl.InspectContainer(id)
	if err != nil {
		return nil, err
	}
	dc.inspectCache[id] = info

	return info, nil
}

// ListContainersAsMap returns list of all available docker containers and base information about them (status, uptime, etc.)
func (dc *DockerClient) ListContainersAsMap() (map[string]docker.APIContainers, error) {
	containers := make(map[string]docker.APIContainers)

	containerList, err := dc.cl.ListContainers(docker.ListContainersOptions{})

	if err != nil {
		return nil, err
	}

	for _, cont := range containerList {
		shortID, err := GetShortID(cont.ID)
		if err != nil {
			return nil, err
		}
		containers[shortID] = cont
	}

	containers["root"] = docker.APIContainers{ID: "/"}

	if len(containers) == 0 {
		return nil, errors.New("No docker container found")
	}

	return containers, nil
}

func getSubsystemPath(subsystem string, id string) (string, error) {
	var subsystemPath string
	systemSlice := "system.slice"
	// hack for finding proper mount point for shares
	// cpu shares are part of cpu group, but openlibcontainers does not support it
	if subsystem == "shares" {
		subsystem = "cpu"
	}
	groupPath, err := cgroups.FindCgroupMountpoint(subsystem)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARNING] Could not find mount point for %v\n", subsystem)
		return "", err
	}
	if isHost(id) {
		if isRunningSystemd() {
			subsystemPath = filepath.Join(groupPath, systemSlice)
		} else {
			subsystemPath = groupPath
		}
		return subsystemPath, nil
	}

	if isFsCgroupParent(groupPath) {
		// default cgroupfs parent is used for container
		subsystemPath = filepath.Join(groupPath, "docker", id)
	} else {
		// cgroup is created under systemd.slice
		subsystemPath = filepath.Join(groupPath, systemSlice, "docker-"+id+".scope")
	}

	return subsystemPath, nil
}

// isFullLengthID returns true if docker ID is a full-length (64 chars)
func isFullLengthID(dockerID string) bool {
	if len(dockerID) == 64 {
		return true
	}

	return false
}

// isFsCgroupParent returns true if the docker was run with default cgroup parent
func isFsCgroupParent(groupPath string) bool {
	fi, err := os.Lstat(filepath.Join(groupPath, "docker"))
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// isRunningSystemd returns true if the host was booted with systemd
func isRunningSystemd() bool {
	fi, err := os.Lstat("/run/systemd/system")
	if err != nil {
		return false
	}

	return fi.IsDir()
}

// isHost returns true if a given id pointing to host
func isHost(id string) bool {
	if id == "/" {
		// it's a host
		return true
	}

	return false
}

// version returns version of docker engine
func (dc *DockerClient) version() (version []int, _ error) {
	version = []int{0, 0}
	env, err := dc.cl.Version()
	if err != nil {
		return version, err
	}
	parseInt := func(str string, defVal int) int {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return defVal
		}
		return int(val)
	}

	for _, kv := range *env {
		kvs := strings.Split(kv, "=")
		if len(kvs) < 2 {
			return nil, fmt.Errorf("Cannot retrive the version of docker engine, is `%v`, expected e.g.`Version = 1.10`", kv)
		}

		if kvs[0] != dockerVersionKey {
			continue
		}

		versionSplit := strings.Split(kvs[1], ".")
		if len(versionSplit) < 2 {
			return nil, fmt.Errorf("Invalid format of docker engine version, is `%v`, expected e.g. `1.10", kvs[1])
		}
		version := []int{parseInt(versionSplit[0], 0), parseInt(versionSplit[1], 0)}
		return version, nil
	}
	return version, nil
}
