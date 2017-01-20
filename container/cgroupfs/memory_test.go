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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

const (
	memoryStatContent = `cache 1111
rss 2222
rss_huge 3333
mapped_file 4444
dirty 5555
writeback 6666
pgpgin 7777
pgpgout 8888
pgfault 9999
pgmajfault 11111
inactive_anon 22222
active_anon 33333
inactive_file 44444
active_file 55555
unevictable 66666
hierarchical_memory_limit 77777
total_cache 88888
total_rss 99999
total_rss_huge 111
total_mapped_file 222
total_dirty 333
total_writeback 444
total_pgpgin 555
total_pgpgout 666
total_pgfault 777
total_pgmajfault 888
total_inactive_anon 999
total_active_anon 11
total_inactive_file 22
total_active_file 33
total_unevictable 44
`
)

type MemorySuite struct {
	suite.Suite
	memoryPath string
}

func (suite *MemorySuite) SetupSuite() {
	suite.memoryPath = "/tmp/memory_test"

	err := os.Mkdir(suite.memoryPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
	modules := []string{"", ".kmem", ".memsw"}
	suite.writeFile(filepath.Join(suite.memoryPath, "memory.stat"), []byte(memoryStatContent))
	for _, module := range modules {
		suite.writeFile(filepath.Join(suite.memoryPath, fmt.Sprintf("memory%s.usage_in_bytes", module)), []byte("111"))
		suite.writeFile(filepath.Join(suite.memoryPath, fmt.Sprintf("memory%s.max_usage_in_bytes", module)), []byte("222"))
		suite.writeFile(filepath.Join(suite.memoryPath, fmt.Sprintf("memory%s.failcnt", module)), []byte("333"))
	}

}

func (suite *MemorySuite) TearDownSuite() {
	err := os.RemoveAll(suite.memoryPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestMemorySuite(t *testing.T) {
	suite.Run(t, &MemorySuite{})
}

func (suite *MemorySuite) TestMemoryGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.memoryPath}
		memory := Memory{}
		err := memory.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(len(stats.Cgroups.MemoryStats.Stats), ShouldEqual, 35)
		So(stats.Cgroups.MemoryStats.Stats["total_mapped_file"], ShouldEqual, 222)
		So(stats.Cgroups.MemoryStats.Stats["inactive_anon"], ShouldEqual, 22222)
		So(stats.Cgroups.MemoryStats.Stats["total_active_anon"], ShouldEqual, 11)
		So(stats.Cgroups.MemoryStats.Stats["rss_huge"], ShouldEqual, 3333)
		So(stats.Cgroups.MemoryStats.Stats["total_writeback"], ShouldEqual, 444)
		So(stats.Cgroups.MemoryStats.Stats["dirty"], ShouldEqual, 5555)
		So(stats.Cgroups.MemoryStats.Stats["working_set"], ShouldEqual, 0)
		So(stats.Cgroups.MemoryStats.Stats["cache"], ShouldEqual, 1111)
	})
}

func (suite *MemorySuite) TestMemoryCacheGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.memoryPath}
		memory := MemoryCache{}
		err := memory.GetStats(stats, opts)
		So(err, ShouldBeNil)
		//So(stats.Cgroups.MemoryStats.Cache, ShouldEqual, 1111)
	})
}

func (suite *MemorySuite) TestMemoryUsageGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.memoryPath}
		memory := MemoryUsage{}
		err := memory.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.MemoryStats.Usage.Usage, ShouldEqual, 111)
		So(stats.Cgroups.MemoryStats.Usage.MaxUsage, ShouldEqual, 222)
		So(stats.Cgroups.MemoryStats.Usage.Failcnt, ShouldEqual, 333)
	})
}

func (suite *MemorySuite) TestMemoryKernelUsageGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.memoryPath}
		memory := KernelMemUsage{}
		err := memory.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.MemoryStats.KernelUsage.Usage, ShouldEqual, 111)
		So(stats.Cgroups.MemoryStats.KernelUsage.MaxUsage, ShouldEqual, 222)
		So(stats.Cgroups.MemoryStats.KernelUsage.Failcnt, ShouldEqual, 333)
	})
}

func (suite *MemorySuite) TestMemorySwapUsageGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.memoryPath}
		memory := SwapMemUsage{}
		err := memory.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.MemoryStats.SwapUsage.Usage, ShouldEqual, 111)
		So(stats.Cgroups.MemoryStats.SwapUsage.MaxUsage, ShouldEqual, 222)
		So(stats.Cgroups.MemoryStats.SwapUsage.Failcnt, ShouldEqual, 333)
	})
}

func (suite *MemorySuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
