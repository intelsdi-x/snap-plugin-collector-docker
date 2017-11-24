// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package container

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/intelsdi-x/snap-plugin-collector-docker/config"
	log "github.com/sirupsen/logrus"
)

const (
	dockerVersionKey string = "Version"
)

// DockerClientInterface provides methods i.a. for interaction with the docker API.
type DockerClientInterface interface {
	ListContainersAsMap() (map[string]*ContainerData, error)
	InspectContainer(string) (*docker.Container, error)
	FindCgroupMountpoint(string, string) (string, error)
	FindControllerMountpoint(string, string, string) (string, error)
	GetDockerParams(...string) (map[string]string, error)
}

// DockerClient holds go-dockerclient instance ready for communication with the server endpoint `unix:///var/run/docker.sock`,
// cache instance which is used to store output from docker container inspect (to avoid execute inspect request multiply times, it is called only once per container)
type DockerClient struct {
	cl           *docker.Client
	inspectCache map[string]*docker.Container
	inspectMutex sync.Mutex
}

type deviceInfo struct {
	device string
	major  string
	minor  string
}

// NewDockerClient returns dockerClient instance ready for communication with the server endpoint `unix:///var/run/docker.sock`
func NewDockerClient(endpoint string) (*DockerClient, error) {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("Cannot initialize docker client instance with the given endpoint `%s`, err=%v", endpoint, err)
	}

	err = client.Ping()
	if err != nil {
		return nil, err
	}

	dc := &DockerClient{
		cl:           client,
		inspectCache: map[string]*docker.Container{},
	}

	config.DockerVersion, err = dc.version()
	if err != nil {
		return nil, err
	}

	return dc, nil
}

// InspectContainer returns details information about running container
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

// GetDockerParam returns given map of parameter/value from running docker engine
func (dc *DockerClient) GetDockerParams(params ...string) (map[string]string, error) {
	env, err := dc.cl.Info()
	if err != nil {
		return nil, err
	}

	vals := make(map[string]string, len(params))

	for _, param := range params {
		if !env.Exists(param) {
			return nil, fmt.Errorf("%s not found", param)
		}
		vals[param] = env.Get(param)
	}

	return vals, nil
}

// GetShortID returns short container ID (12 chars)
func GetShortID(dockerID string) (string, error) {
	if dockerID == "root" {
		return dockerID, nil
	}

	if len(dockerID) < 12 {
		return "", fmt.Errorf("Docker id %v is too short (the length of id should equal at least 12)", dockerID)
	}

	return dockerID[:12], nil
}

// ListContainersAsMap returns list of all available docker containers and base information about them (status, uptime, etc.)
func (dc *DockerClient) ListContainersAsMap() (map[string]*ContainerData, error) {
	containers := make(map[string]*ContainerData)

	containerList, err := dc.cl.ListContainers(docker.ListContainersOptions{})

	if err != nil {
		return nil, err
	}

	for _, c := range containerList {
		shortID, err := GetShortID(c.ID)
		if err != nil {
			return nil, err
		}

		spec := Specification{
			Status:     c.Status,
			Created:    time.Unix(c.Created, 0).Format("2006-01-02T15:04:05Z07:00"),
			Image:      c.Image,
			SizeRw:     c.SizeRw,
			SizeRootFs: c.SizeRootFs,
			Labels:     c.Labels,
		}

		containerData := ContainerData{
			ID:            c.ID,
			Specification: spec,
			Stats:         NewStatistics(),
		}

		containers[shortID] = &containerData
	}

	if len(containers) == 0 {
		log.WithFields(log.Fields{
			"block":    "client",
			"function": "ListContainersAsMap",
		}).Warnf("no running containers on host")
	}

	containers["root"] = &ContainerData{ID: "/", Stats: NewStatistics()}

	return containers, nil
}

// FindCgroupMountpoint returns cgroup mountpoint of a given subsystem
func (dc *DockerClient) FindCgroupMountpoint(procfs string, subsystem string) (string, error) {
	f, err := os.Open(filepath.Join(procfs, "self/mountinfo"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("Cgroup {%s} mountpoint not found", subsystem)
}

// FindControllerMountpoint returns mountpoints of a given controller and container PID
func (dc *DockerClient) FindControllerMountpoint(subsystem, pid, procfs string) (string, error) {
	f, err := os.Open(filepath.Join(procfs, pid, "mountinfo"))
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Fields(txt)
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return filepath.Join(filepath.Dir(fields[4]), subsystem, fields[3]), nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("can't find mountpoint for controller {%s} for container pid {%s}", subsystem, pid)

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
