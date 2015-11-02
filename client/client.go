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
	"fmt"

	"github.com/fsouza/go-dockerclient"
	"github.com/opencontainers/runc/libcontainer/cgroups"
)

const endpoint string = "unix:///var/run/docker.sock"

type ContainerInfo struct {
	Id string
	Status string
	Created int64
	Image string
}

type DockerClientInterface interface {
	ListContainers() ([]ContainerInfo, error)
	FindCgroupMountpoint(subsystem string) (string, error)
}

type dockerClient struct {}

func NewDockerClient() *dockerClient {
	return new(dockerClient)
}

func (dc *dockerClient) FindCgroupMountpoint(subsystem string) (string, error){
	return cgroups.FindCgroupMountpoint(subsystem)
}

func (dc *dockerClient) ListContainers() ([]ContainerInfo, error) {
	cl, err := docker.NewClient(endpoint)
	if err != nil {
		fmt.Println("[ERROR] Could not create docker client!")
		return []ContainerInfo{}, err
	}

	containerList, err := cl.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		fmt.Println("[ERROR] Could not retrieve list of running containers!")
		return []ContainerInfo{}, err
	}

	containers := []ContainerInfo{}
	for _, cont := range containerList {
		cinfo := ContainerInfo{
			Id: cont.ID,
			Status: cont.Status,
			Created: cont.Created,
			Image: cont.Image}
		containers = append(containers, cinfo)
	}
	return containers, nil
}

