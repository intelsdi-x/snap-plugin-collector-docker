// +build linux

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

package cgroupfs

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/docker/go-units"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

var hpControlDir = "/sys/kernel/mm/hugepages"

// HugeTlb implements StatGetter interface
type HugeTlb struct{}

// GetStats reads huge table metrics from Hugetlb Group
func (h *HugeTlb) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	hugePageSizes, err := getHugePageSize(hpControlDir)
	if err != nil {
		return err
	}

	for _, pageSize := range hugePageSizes {
		usage, err := parseIntValue(filepath.Join(path, strings.Join([]string{"hugetlb", pageSize, "usage_in_bytes"}, ".")))
		if err != nil {
			return err
		}

		maxUsage, err := parseIntValue(filepath.Join(path, strings.Join([]string{"hugetlb", pageSize, "max_usage_in_bytes"}, ".")))
		if err != nil {
			return err
		}

		failcnt, err := parseIntValue(filepath.Join(path, strings.Join([]string{"hugetlb", pageSize, "failcnt"}, ".")))
		if err != nil {
			return err
		}

		stats.Cgroups.HugetlbStats[pageSize] = container.HugetlbStats{
			Usage:    usage,
			MaxUsage: maxUsage,
			Failcnt:  failcnt,
		}
	}

	return nil
}

func getHugePageSize(controlDir string) ([]string, error) {
	var pageSizes []string
	sizeList := []string{"B", "kB", "MB", "GB", "TB", "PB"}
	files, err := ioutil.ReadDir(controlDir)
	if err != nil {
		return pageSizes, err
	}
	for _, st := range files {
		nameArray := strings.Split(st.Name(), "-")
		pageSize, err := units.RAMInBytes(nameArray[1])
		if err != nil {
			return nil, err
		}
		sizeString := units.CustomSize("%g%s", float64(pageSize), 1024.0, sizeList)
		pageSizes = append(pageSizes, sizeString)
	}

	return pageSizes, nil
}
