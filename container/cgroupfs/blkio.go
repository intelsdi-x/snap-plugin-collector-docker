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

This file incorporates work covered by the following copyright and permission notice:
	Copyright 2014 Docker, Inc.
Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package contains code from OCI/opencontainers (https://github.com/opencontainers/runc) with following:
// - function getCFQStats
// - function getStats
// - function splitBlkioStatLine
// - function getBlkioStat

package cgroupfs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

// Blkio implements StatGetter interface
type Blkio struct{}

// GetStats reads blkio metrics from Blkio Group from blkio.*
func (b *Blkio) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}
	// Try to read CFQ stats available on all CFQ enabled kernels first
	if blkioStats, err := getBlkioStat(filepath.Join(path, "blkio.io_serviced_recursive")); err == nil && blkioStats != nil {
		return getCFQStats(path, stats)
	}
	return getStats(path, stats) // Use generic stats as fallback
}

func splitBlkioStatLine(r rune) bool {
	return r == ' ' || r == ':'
}

func getBlkioStat(path string) ([]container.BlkioStatEntry, error) {
	var blkioStats []container.BlkioStatEntry
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return blkioStats, nil
		}
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// format: dev type amount
		fields := strings.FieldsFunc(sc.Text(), splitBlkioStatLine)
		if len(fields) < 3 {
			if len(fields) == 2 && fields[0] == "Total" {
				// skip total line
				continue
			} else {
				return nil, fmt.Errorf("Invalid line found while parsing %s: %s", path, sc.Text())
			}
		}

		major, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return nil, err
		}

		minor, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return nil, err
		}

		op := ""
		index := 2
		if len(fields) == 4 {
			op = fields[2]
			index = 3
		}
		val, err := strconv.ParseUint(fields[index], 10, 64)
		if err != nil {
			return nil, err
		}
		blkioStats = append(blkioStats, container.BlkioStatEntry{Major: major, Minor: minor, Op: op, Value: val})
	}

	return blkioStats, nil
}

func getCFQStats(path string, stats *container.Statistics) error {
	var blkioStats []container.BlkioStatEntry
	var err error

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.sectors_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.SectorsRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_service_bytes_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoServiceBytesRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_serviced_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoServicedRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_queued_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoQueuedRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_service_time_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoServiceTimeRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_wait_time_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoWaitTimeRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.io_merged_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoMergedRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.time_recursive")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoTimeRecursive = blkioStats

	return nil
}

func getStats(path string, stats *container.Statistics) error {
	var blkioStats []container.BlkioStatEntry
	var err error

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.throttle.io_service_bytes")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoServiceBytesRecursive = blkioStats

	if blkioStats, err = getBlkioStat(filepath.Join(path, "blkio.throttle.io_serviced")); err != nil {
		return err
	}
	stats.Cgroups.BlkioStats.IoServicedRecursive = blkioStats

	return nil
}
