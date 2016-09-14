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

package mocks

import (
	dock "github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/mock"

	"github.com/intelsdi-x/snap-plugin-collector-docker/client"
	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
)

type ClientMock struct {
	mock.Mock
}

func (cm *ClientMock) FindCgroupMountpoint(subsystem string) (string, error) {
	ret := cm.Called(subsystem)

	return ret.String(0), ret.Error(1)
}

func (cm *ClientMock) NewDockerClient() (*client.DockerClient, error) {
	args := cm.Called()

	var r0 *client.DockerClient
	if args.Get(0) != nil {
		r0 = args.Get(0).(*client.DockerClient)
	}
	return r0, args.Error(1)
}

func (cm *ClientMock) ListContainersAsMap() (map[string]dock.APIContainers, error) {
	args := cm.Called()

	var r0 map[string]dock.APIContainers
	if args.Get(0) != nil {
		r0 = args.Get(0).(map[string]dock.APIContainers)
	}
	return r0, args.Error(1)
}

func (cm *ClientMock) InspectContainer(string) (*dock.Container, error) {
	args := cm.Called()

	var r0 *dock.Container

	if args.Get(0) != nil {
		r0 = args.Get(0).(*dock.Container)
	}

	return r0, args.Error(1)
}

func (cm *ClientMock) GetStatsFromContainer(string, bool) (*wrapper.Statistics, error) {
	args := cm.Called()

	var r0 *wrapper.Statistics
	if args.Get(0) != nil {
		r0 = args.Get(0).(*wrapper.Statistics)
	}
	return r0, args.Error(1)
}

/*
type StatsMock struct {
	mock.Mock
}

func (s *StatsMock) GetStats(path string, stats *cgroups.Stats) error {
	ret := s.Mock.Called(path, stats)
	return ret.Error(0)
}
*/

func CreateMockStats() *wrapper.Statistics {
	stats := wrapper.NewStatistics()
	stats.Connection.Tcp.Close = 1
	stats.Connection.Tcp.Established = 1

	stats.Connection.Tcp6.Close = 2
	stats.Connection.Tcp6.Established = 2

	stats.Network = []wrapper.NetworkInterface{
		wrapper.NetworkInterface{
			Name:      "eth0",
			RxBytes:   1024,
			RxPackets: 1,
			TxBytes:   4096,
			TxPackets: 4,
			TxErrors:  0,
		},
		wrapper.NetworkInterface{
			Name:      "eth1",
			RxBytes:   2048,
			RxPackets: 2,
			TxBytes:   0,
			TxPackets: 0,
			TxErrors:  10,
		},
	}
	stats.Filesystem["dev1"] = wrapper.FilesystemInterface{
		Device:         "dev1",
		Type:           "vfs",
		Limit:          0,
		Usage:          0,
		BaseUsage:      0,
		Available:      0,
		InodesFree:     0,
		ReadsCompleted: 0,
	}

	stats.Filesystem["dev2"] = wrapper.FilesystemInterface{
		Device:         "dev2",
		Type:           "vfs",
		Limit:          0,
		Usage:          0,
		BaseUsage:      0,
		Available:      0,
		InodesFree:     0,
		ReadsCompleted: 0,
	}

	stats.CgroupStats.CpuStats.CpuUsage.PercpuUsage = []uint64{1000, 2000, 3000, 4000}

	return stats
}
