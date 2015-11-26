// +build unit

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
	"github.com/stretchr/testify/mock"

	"github.com/opencontainers/runc/libcontainer/cgroups"

	. "github.com/intelsdi-x/snap-plugin-collector-docker/client"
	_ "github.com/intelsdi-x/snap-plugin-collector-docker/tools"
)

type ClientMock struct {
	mock.Mock
}

func (cm *ClientMock) ListContainers() ([]ContainerInfo, error) {
	ret := cm.Mock.Called()
	return ret.Get(0).([]ContainerInfo), ret.Error(1)
}

func (cm *ClientMock) FindCgroupMountpoint(subsystem string) (string, error) {
	ret := cm.Mock.Called(subsystem)
	return ret.String(0), ret.Error(1)
}

type StatsMock struct {
	mock.Mock
}

func (s *StatsMock) GetStats(path string, stats *cgroups.Stats) error {
	ret := s.Mock.Called(path, stats)
	return ret.Error(0)
}

type ToolsMock struct {
	mock.Mock
}

func (tm *ToolsMock) Map2Namespace(stats map[string]interface{}, current string, out *[]string) {
	_ = tm.Mock.Called(stats, current, out)
}

func (tm *ToolsMock) GetValueByField(object interface{}, fields []string) interface{} {
	ret := tm.Mock.Called(object, fields)
	return ret.Get(0).(interface{})
}

func (tm *ToolsMock) GetValueByNamespace(object interface{}, ns []string) interface{} {
	ret := tm.Mock.Called(object, ns)
	return ret.Get(0).(interface{})
}
