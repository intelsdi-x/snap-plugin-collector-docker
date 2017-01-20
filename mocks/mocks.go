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
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/mock"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

type ClientMock struct {
	mock.Mock
}

func (cm *ClientMock) FindCgroupMountpoint(procfs string, subsystem string) (string, error) {
	ret := cm.Called(procfs, subsystem)

	return ret.String(0), ret.Error(1)
}

func (cm *ClientMock) FindControllerMountpoint(cgroupPath, pid, procfs string) (string, error) {
	ret := cm.Called(cgroupPath, pid, procfs)
	return ret.String(0), ret.Error(1)
}

func (cm *ClientMock) NewDockerClient() (*container.DockerClient, error) {
	args := cm.Called()

	var r0 *container.DockerClient
	if args.Get(0) != nil {
		r0 = args.Get(0).(*container.DockerClient)
	}
	return r0, args.Error(1)
}

func (cm *ClientMock) ListContainersAsMap() (map[string]*container.ContainerData, error) {
	args := cm.Called()

	var r0 map[string]*container.ContainerData
	if args.Get(0) != nil {
		r0 = args.Get(0).(map[string]*container.ContainerData)
	}
	return r0, args.Error(1)
}

func (cm *ClientMock) InspectContainer(string) (*docker.Container, error) {
	args := cm.Called()

	var r0 *docker.Container

	if args.Get(0) != nil {
		r0 = args.Get(0).(*docker.Container)
	}

	return r0, args.Error(1)
}

func (cm *ClientMock) GetDockerParams(params ...string) (map[string]string, error) {
	ret := cm.Called(params)
	return ret.Get(0).(map[string]string), ret.Error(1)
}

var MockGetters map[string]container.StatGetter = map[string]container.StatGetter{
	"cpu_usage":  &MockCpuAcct{},
	"cache":      &MockMemCache{},
	"usage":      &MockMemUsage{},
	"statistics": &MockMemStats{},
	"network":    &MockNet{},
	"tcp":        &MockTcp{},
	"tcp6":       &MockTcp{},
}

type MockCpuAcct struct{}

func (m *MockCpuAcct) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Cgroups.CpuStats.CpuUsage.PerCpu = []uint64{1111, 2222, 3333, 4444}
	return nil
}

type MockMemCache struct{}

func (m *MockMemCache) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Cgroups.MemoryStats.Cache = 1111
	return nil
}

type MockMemUsage struct{}

func (m *MockMemUsage) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Cgroups.MemoryStats.Usage.Failcnt = 1111
	stats.Cgroups.MemoryStats.Usage.Usage = 2222
	stats.Cgroups.MemoryStats.Usage.MaxUsage = 3333
	return nil
}

type MockMemStats struct{}

func (m *MockMemStats) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Cgroups.MemoryStats.Stats = map[string]uint64{"pgpgin": 11111}
	return nil
}

type MockNet struct{}

func (m *MockNet) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Network = []container.NetworkInterface{{Name: "eth0", TxBytes: 1111, RxBytes: 2222}}
	return nil
}

type MockTcp struct{}

func (m *MockTcp) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	stats.Connection.Tcp.Established = 1111
	return nil
}
