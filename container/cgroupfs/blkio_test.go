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

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

const (
	blkioContents = `8:0 Read 100
8:0 Write 200
8:0 Sync 300
8:0 Async 500
8:0 Total 500
Total 500`
)

type BlkioSuite struct {
	suite.Suite
	blkioPath string
}

func TestBlkioSuite(t *testing.T) {
	suite.Run(t, &BlkioSuite{})
}

func (s *BlkioSuite) SetupSuite() {
	s.blkioPath = "/tmp/blkio_test"
	err := os.Mkdir(s.blkioPath, 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_service_bytes_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_serviced_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_queued_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_service_time_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_wait_time_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.io_merged_recursive"), []byte(blkioContents))
	s.writeFile(filepath.Join(s.blkioPath, "blkio.time_recursive"), []byte(blkioContents))
}

func (s *BlkioSuite) TearDownSuite() {
	err := os.RemoveAll(s.blkioPath)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *BlkioSuite) TestgetBlkioStat() {
	testCases := []testCase{
		{ExpectedMajor: 8, ExpectedMinor: 0, ExpectedValue: 100, ExpectedOp: "Read"},
		{ExpectedMajor: 8, ExpectedMinor: 0, ExpectedValue: 200, ExpectedOp: "Write"},
		{ExpectedMajor: 8, ExpectedMinor: 0, ExpectedValue: 300, ExpectedOp: "Sync"},
		{ExpectedMajor: 8, ExpectedMinor: 0, ExpectedValue: 500, ExpectedOp: "Async"},
		{ExpectedMajor: 8, ExpectedMinor: 0, ExpectedValue: 500, ExpectedOp: "Total"},
	}

	Convey("Read blkio.io_service_bytes_recursive content", s.T(), func() {
		blkio, err := getBlkioStat(filepath.Join(s.blkioPath, "blkio.io_service_bytes_recursive"))
		So(err, ShouldBeNil)
		So(len(blkio), ShouldEqual, 5)
		for i := 0; i < len(testCases); i++ {
			So(blkio[i].Major, ShouldEqual, testCases[i].ExpectedMajor)
			So(blkio[i].Minor, ShouldEqual, testCases[i].ExpectedMinor)
			So(blkio[i].Value, ShouldEqual, testCases[i].ExpectedValue)
			So(blkio[i].Op, ShouldEqual, testCases[i].ExpectedOp)
		}
	})
}

func (s *BlkioSuite) TestGetStatsPositive() {
	Convey("Call GetStats", s.T(), func() {
		blkio := Blkio{}
		stats := container.NewStatistics()
		err := blkio.GetStats(stats, container.GetStatOpt{"cgroup_path": s.blkioPath})
		So(err, ShouldBeNil)
		So(len(stats.Cgroups.BlkioStats.IoMergedRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoQueuedRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoServiceBytesRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoServicedRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoServiceTimeRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoTimeRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.IoWaitTimeRecursive), ShouldEqual, 5)
		So(len(stats.Cgroups.BlkioStats.SectorsRecursive), ShouldEqual, 0)
	})
}

func (s *BlkioSuite) TestGetStatsNegative() {
	Convey("Call GetStats", s.T(), func() {
		blkio := Blkio{}
		err := blkio.GetStats(container.NewStatistics(), container.GetStatOpt{})
		So(err, ShouldNotBeNil)
	})
}

type testCase struct {
	ExpectedMajor uint64
	ExpectedMinor uint64
	ExpectedValue uint64
	ExpectedOp    string
}

func (s *BlkioSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		s.T().Fatal(err)
	}
}
