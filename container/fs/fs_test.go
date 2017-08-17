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

This file incorporates work covered by the following copyright and permission notice:
	Copyright 2014 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/moby/moby/pkg/mount"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

const (
	procContents         = `8      1 sdd2 40 0 280 223 7 0 22 108 0 330 330`
	aufsContents         = `0      42 sdd2 40 0 280 223 7 0 22 108 0 330 330`
	overlayContents      = `8      1 sdd2 40 0 280 223 7 0 22 108 0 330 330`
	usageContent         = `6379736 /tmp/usage`
	usageContentInvalid  = `six /tmp/usage`
	invalidContents      = `8      1 sdd2 40 0 280 223 7 0 108 0 330 330`
	invalidContentsParse = `a      1 sdd2 40 0 280 223 7 0 22 108 0 330 330`
)

type FsSuite struct {
	suite.Suite
	fsPath string
}

func TestFsSuite(t *testing.T) {
	suite.Run(t, &FsSuite{})
}

func (s *FsSuite) SetupSuite() {
	err := os.MkdirAll("/tmp/aufs/diff/27fa0900fe22", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/overlay/27fa0900fe22", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.Mkdir("/tmp/overlay/root", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.Mkdir("/tmp/aufs/root", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.Mkdir("/tmp/proc", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/var/lib/docker/containers/27fa0900fe22", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/var/lib/docker/aufs/diff", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/var/lib/docker/overlay", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/var/lib/docker/zfs", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/var/lib/docker/zfs", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.MkdirAll("/tmp/test/dev/sdf100", 0700)
	if err != nil {
		s.T().Fatal(err)
	}
	s.writeFile(filepath.Join("/tmp/test", "usage"), []byte(usageContent))
	s.writeFile(filepath.Join("/tmp/test", "usage_invalid_parse"), []byte(usageContentInvalid))
	s.writeFile(filepath.Join("/tmp/proc", "diskstats"), []byte(procContents))
	s.writeFile(filepath.Join("/tmp/proc", "diskstat_invalid_1"), []byte(invalidContents))
	s.writeFile(filepath.Join("/tmp/proc", "diskstat_invalid_2"), []byte(invalidContentsParse))
	s.writeFile(filepath.Join("/tmp/overlay/27fa0900fe22", "diskstats"), []byte(overlayContents))
	s.writeFile(filepath.Join("/tmp/aufs/diff/27fa0900fe22", "diskstats"), []byte(aufsContents))
	buf := new(syscall.Stat_t)
	err = syscall.Stat("/tmp/test", buf)
	if err != nil {
		s.T().Fatal(err)
	}
	mockValidMount := mount.Info{
		ID:         250,
		Parent:     0,
		Major:      int(major(buf.Dev)),
		Minor:      int(minor(buf.Dev)),
		Root:       "/tmp/test",
		Mountpoint: "/",
		Opts:       "rw,relatime",
		Optional:   "shared:1",
		Fstype:     "ext4",
		Source:     "/tmp/test/dev/sdf100",
		VfsOpts:    "rw,errors=remount-ro,data=ordered",
	}
	mockedMounts = append(mockedMounts, &mockValidMount)
}

func (s *FsSuite) TearDownSuite() {
	err := os.RemoveAll("/tmp/aufs/")
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.RemoveAll("/tmp/overlay/")
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.RemoveAll("/tmp/proc/")
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.RemoveAll("/tmp/var/")
	if err != nil {
		s.T().Fatal(err)
	}
	err = os.RemoveAll("/tmp/test/")
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *FsSuite) TestFS() {
	Convey("Check usage parser valid options", s.T(), func() {
		usage, err := diskUsage("cat", []string{"/tmp/test/usage"})
		So(usage, ShouldEqual, 6379736)
		So(err, ShouldBeNil)
	})
	Convey("Check usage parser invalid options", s.T(), func() {
		usage, err := diskUsage("cat", []string{"/tmp/test/usage_invalid"})
		So(usage, ShouldEqual, 0)
		So(err, ShouldNotBeNil)
	})
	Convey("Check usage parser invalid data", s.T(), func() {
		usage, err := diskUsage("cat", []string{"/tmp/test/usage_invalid_parse"})
		So(usage, ShouldEqual, 0)
		So(err, ShouldNotBeNil)
	})
	Convey("Check getDiskStatsMap", s.T(), func() {
		dsm, err := getDiskStatsMap("/tmp/proc/diskstats")
		So(err, ShouldBeNil)
		So(dsm[DeviceId{Major: 8, Minor: 1}].ReadsCompleted, ShouldEqual, 40)
		So(dsm[DeviceId{Major: 8, Minor: 1}].ReadsMerged, ShouldEqual, 0)
		So(dsm[DeviceId{Major: 8, Minor: 1}].SectorsRead, ShouldEqual, 280)
		So(dsm[DeviceId{Major: 8, Minor: 1}].ReadTime, ShouldEqual, 223)
		So(dsm[DeviceId{Major: 8, Minor: 1}].WritesCompleted, ShouldEqual, 7)
		So(dsm[DeviceId{Major: 8, Minor: 1}].WritesMerged, ShouldEqual, 0)
		So(dsm[DeviceId{Major: 8, Minor: 1}].SectorsWritten, ShouldEqual, 22)
		So(dsm[DeviceId{Major: 8, Minor: 1}].WriteTime, ShouldEqual, 108)
		So(dsm[DeviceId{Major: 8, Minor: 1}].IoInProgress, ShouldEqual, 0)
		So(dsm[DeviceId{Major: 8, Minor: 1}].IoTime, ShouldEqual, 330)
		So(dsm[DeviceId{Major: 8, Minor: 1}].WeightedIoTime, ShouldEqual, 330)
	})
	Convey("Check getDiskStatsMap", s.T(), func() {
		_, err := getDiskStatsMap("/tmp/proc/diskstat")
		So(err, ShouldBeNil)
	})
	Convey("Check getDiskStatsMap invalid file content (too short)", s.T(), func() {
		dsm, err := getDiskStatsMap("/tmp/proc/diskstat_invalid_1")
		So(dsm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
	Convey("Check getDiskStatsMap invalid file content (non numeric value)", s.T(), func() {
		dsm, err := getDiskStatsMap("/tmp/proc/diskstat_invalid_2")
		So(dsm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
	Convey("FS tests", s.T(), func() {
		fsInfo, err := newFsInfo("test")
		So(fsInfo, ShouldNotBeNil)
		So(err, ShouldBeNil)
		Convey("GetDirFsDevice", func() {
			Convey("Valid path", func() {
				dir, err := fsInfo.GetDirFsDevice("/tmp/test")
				So(dir, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("Invalid path", func() {
				dir, err := fsInfo.GetDirFsDevice("/invalid")
				So(dir, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
		Convey("GetGlobalFsInfo", func() {
			Convey("Valid path", func() {
				_, err := fsInfo.GetGlobalFsInfo("/tmp/proc/")
				So(err, ShouldBeNil)
			})
		})

	})
	usage := make(map[string]uint64)
	du := DiskUsageCollector{DiskUsage: usage}
	Convey("GetStats tests", s.T(), func() {
		du.DiskUsage["/tmp"] = 2000
		du.DiskUsage["/tmp/aufs/diff/27fa0900fe22"] = 250
		du.DiskUsage["/tmp/overlay/27fa0900fe22"] = 929
		Convey("Check invalid root dir", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "27fa0900fe22", "container_drv": "aufs", "procfs": "/tmp/proc", "root_dir": "/invalid_dir"})
			// Even if root_dir is invalid there is no error returned.
			// But there is logged message started with: `Os.Stat failed`
			So(err, ShouldBeNil)
		})
		Convey("Check valid aufs driver", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "27fa0900fe22", "container_drv": "aufs", "procfs": "/tmp/proc", "root_dir": "/tmp"})
			So(err, ShouldBeNil)
		})
		Convey("Check valid overlay driver", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "27fa0900fe22", "container_drv": "overlay", "procfs": "/tmp/proc", "root_dir": "/tmp"})
			So(err, ShouldBeNil)
		})
		Convey("Check invalid driver", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "27fa0900fe22", "container_drv": "ext", "procfs": "/tmp/proc", "root_dir": "/tmp"})
			So(err, ShouldNotBeNil)
		})
		Convey("Check root valid aufs driver", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "root", "container_drv": "aufs", "procfs": "/tmp/proc", "root_dir": "/tmp"})
			So(err, ShouldBeNil)
		})
		Convey("Check root valid overlay driver", func() {
			stats := container.NewStatistics()
			err := du.GetStats(stats, container.GetStatOpt{"container_id": "root", "container_drv": "overlay", "procfs": "/tmp/proc", "root_dir": "/tmp"})
			So(err, ShouldBeNil)
		})
	})
}

func (s *FsSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		s.T().Fatal(err)
	}
}
