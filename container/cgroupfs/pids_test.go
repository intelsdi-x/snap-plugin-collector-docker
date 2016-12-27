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

type PidsSuite struct {
	suite.Suite
	pidsPath string
}

func (suite *PidsSuite) SetupSuite() {
	suite.pidsPath = "/tmp/pids_test"
	err := os.Mkdir(suite.pidsPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.writeFile(filepath.Join(suite.pidsPath, "pids.current"), []byte("1"))
	suite.writeFile(filepath.Join(suite.pidsPath, "pids.max"), []byte("2"))

}

func (suite *PidsSuite) TearDownSuite() {
	err := os.RemoveAll(suite.pidsPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestPidsSuite(t *testing.T) {
	suite.Run(t, &PidsSuite{})
}

func (suite *PidsSuite) TestPidsGetStats() {
	Convey("collecting data from cpuset controller", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.pidsPath}
		pids := Pids{}
		err := pids.GetStats(stats, opts)
		So(err, ShouldBeNil)
		So(stats.Cgroups.PidsStats.Current, ShouldEqual, 1)
		So(stats.Cgroups.PidsStats.Limit, ShouldEqual, 2)
	})
}

func (suite *PidsSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
