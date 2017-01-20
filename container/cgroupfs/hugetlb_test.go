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
	hugetlbUsageContents    = "128\n"
	hugetlbMaxUsageContents = "256\n"
	hugetlbFailcnt          = "100\n"

	hpSize1GB = "hugepages-1048576kB"
	hpSize1MB = "hugepages-2048kB"
)

var (
	usage    = "hugetlb.%s.usage_in_bytes"
	maxUsage = "hugetlb.%s.max_usage_in_bytes"
	failcnt  = "hugetlb.%s.failcnt"
)

type HugePagesSuite struct {
	suite.Suite
	hugepagesPath string
	pageSizes     []string
}

func (suite *HugePagesSuite) SetupSuite() {
	suite.hugepagesPath = "/tmp/hugepages_test"
	hpControlDir = filepath.Join(suite.hugepagesPath, "control_dir")

	err := os.Mkdir(suite.hugepagesPath, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
	err = os.Mkdir(hpControlDir, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
	err = os.Mkdir(filepath.Join(suite.hugepagesPath, hpSize1MB), 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
	err = os.Mkdir(filepath.Join(suite.hugepagesPath, hpSize1GB), 0700)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.pageSizes = []string{"1GB", "2MB"}
	for _, pageSize := range suite.pageSizes {
		suite.writeFile(filepath.Join(suite.hugepagesPath, fmt.Sprintf(usage, pageSize)), []byte(hugetlbUsageContents))
		suite.writeFile(filepath.Join(suite.hugepagesPath, fmt.Sprintf(maxUsage, pageSize)), []byte(hugetlbMaxUsageContents))
		suite.writeFile(filepath.Join(suite.hugepagesPath, fmt.Sprintf(failcnt, pageSize)), []byte(hugetlbFailcnt))
	}
}

func (suite *HugePagesSuite) TearDownSuite() {
	err := os.RemoveAll(suite.hugepagesPath)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func TestHugePagesSuite(t *testing.T) {
	suite.Run(t, &HugePagesSuite{})
}

func (suite *HugePagesSuite) TestGetStats() {
	Convey("", suite.T(), func() {
		stats := container.NewStatistics()
		opts := container.GetStatOpt{"cgroup_path": suite.hugepagesPath}
		hugetlb := HugeTlb{}
		err := hugetlb.GetStats(stats, opts)
		So(err, ShouldBeNil)
	})
}

func (suite *HugePagesSuite) writeFile(path string, content []byte) {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		suite.T().Fatal(err)
	}
}
