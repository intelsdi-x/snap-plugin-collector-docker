// +build small

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

package cgroupfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

type CpuSetSuite struct {
	suite.Suite
	cpusetPath string
}

func (suite *CpuSetSuite) SetupSuite() {
	suite.cpusetPath = "/tmp/cpuset_test"
	err := os.Mkdir(suite.cpusetPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.memory_migrate"), []byte("1"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.cpu_exclusive"), []byte("2"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.mem_exclusive"), []byte("3"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.mems"), []byte("4"))
	suite.writeFile(filepath.Join(suite.cpusetPath, "cpuset.cpus"), []byte("5"))
}

func (suite *CpuSetSuite) TearDownSuite() {
	err := os.RemoveAll(suite.cpusetPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestCpuSetSuite(t *testing.T) {
	suite.Run(t, &CpuSetSuite{})
}

func (suite *CpuSetSuite) TestCpuGetStats() {
	Convey("collecting data from cpuset controller", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.cpusetPath}
		cpu := CpuSet{}
		err := cpu.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.CpuSetStats.MemoryMigrate, ShouldEqual, 1)
		So(stats.Cgroups.CpuSetStats.CpuExclusive, ShouldEqual, 2)
		So(stats.Cgroups.CpuSetStats.MemoryExclusive, ShouldEqual, 3)
		So(stats.Cgroups.CpuSetStats.Mems, ShouldEqual, "4")
		So(stats.Cgroups.CpuSetStats.Cpus, ShouldEqual, "5")
	})
}

func (suite *CpuSetSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
