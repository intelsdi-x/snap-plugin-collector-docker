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

const (
	cpuStatContents = `nr_periods 11
nr_throttled 22
throttled_time 33
`
	cpuAcctStatContents = `user 11111111
system 22222222
`
)

type CpuSuite struct {
	suite.Suite
	cpuPath string
}

func (suite *CpuSuite) SetupSuite() {
	suite.cpuPath = "/tmp/cpu_test"
	err := os.Mkdir(suite.cpuPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.writeFile(filepath.Join(suite.cpuPath, "cpu.stat"), []byte(cpuStatContents))
	suite.writeFile(filepath.Join(suite.cpuPath, "cpuacct.stat"), []byte(cpuAcctStatContents))
	suite.writeFile(filepath.Join(suite.cpuPath, "cpuacct.usage"), []byte("3333333333"))
	suite.writeFile(filepath.Join(suite.cpuPath, "cpuacct.usage_percpu"), []byte("44444444 555555555"))
	suite.writeFile(filepath.Join(suite.cpuPath, "cpu.shares"), []byte("6666"))
}

func (suite *CpuSuite) TearDownSuite() {
	err := os.RemoveAll(suite.cpuPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestCpuSuite(t *testing.T) {
	suite.Run(t, &CpuSuite{})
}

func (suite *CpuSuite) TestCpuGetStats() {
	Convey("collecting data from cpu.stat", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.cpuPath}
		cpu := Cpu{}
		err := cpu.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.CpuStats.ThrottlingData.NrPeriods, ShouldEqual, 11)
		So(stats.Cgroups.CpuStats.ThrottlingData.NrThrottled, ShouldEqual, 22)
		So(stats.Cgroups.CpuStats.ThrottlingData.ThrottledTime, ShouldEqual, 33)

	})
}

func (suite *CpuSuite) TestCpuAcctGetStats() {
	Convey("collecting data from cpuacct.stat", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.cpuPath}
		cpu := CpuAcct{}
		err := cpu.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.CpuStats.CpuUsage.UserMode, ShouldEqual, 11111111)
		So(stats.Cgroups.CpuStats.CpuUsage.KernelMode, ShouldEqual, 22222222)
		So(stats.Cgroups.CpuStats.CpuUsage.Total, ShouldEqual, 3333333333)
		So(stats.Cgroups.CpuStats.CpuUsage.PerCpu[0], ShouldEqual, 44444444)
		So(stats.Cgroups.CpuStats.CpuUsage.PerCpu[1], ShouldEqual, 555555555)
	})
}

func (suite *CpuSuite) TestCpuSharesGetStats() {
	Convey("collecting data from cpu.shares", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.cpuPath}
		cpu := CpuShares{}
		err := cpu.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.CpuStats.CpuShares, ShouldEqual, 6666)
	})
}

func (suite *CpuSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
