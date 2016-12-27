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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

// Cpu implements StatGetter interface
type Cpu struct{}

// GetStats reads throttling metrics from Cpu Group from cpu.stat
func (cpu *Cpu) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(path, "cpu.stat"))
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		param, value, err := parseEntry(scan.Text())
		if err != nil {
			return err
		}

		switch param {
		case "nr_periods":
			stats.Cgroups.CpuStats.ThrottlingData.NrPeriods = value
		case "nr_throttled":
			stats.Cgroups.CpuStats.ThrottlingData.NrThrottled = value
		case "throttled_time":
			stats.Cgroups.CpuStats.ThrottlingData.ThrottledTime = value
		default:
			return fmt.Errorf("Unknown cpu.stat parameter: %s", param)
		}
	}

	return nil
}

// CpuAcct implements StatGetter interface
type CpuAcct struct{}

// GetStats reads usage metrics from Cpu Group from cpuacct.stat
func (cpuacct *CpuAcct) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(path, "cpuacct.stat"))
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		param, value, err := parseEntry(scan.Text())
		if err != nil {
			return err
		}

		switch param {
		case "user":
			stats.Cgroups.CpuStats.CpuUsage.UserMode = value
		case "system":
			stats.Cgroups.CpuStats.CpuUsage.KernelMode = value
		default:
			return fmt.Errorf("Unknown cpuacct.stat parameter: %s", param)
		}
	}

	usages, err := ioutil.ReadFile(filepath.Join(path, "cpuacct.usage_percpu"))
	if err != nil {
		return err
	}

	perCpu := []uint64{}
	for _, usage := range strings.Fields(string(usages)) {
		value, err := strconv.ParseUint(usage, 10, 64)
		if err != nil {
			return err
		}
		perCpu = append(perCpu, value)
	}
	stats.Cgroups.CpuStats.CpuUsage.PerCpu = perCpu

	total, err := parseIntValue(filepath.Join(path, "cpuacct.usage"))
	if err != nil {
		return err
	}
	stats.Cgroups.CpuStats.CpuUsage.Total = total

	return nil
}

// CpuShares implements StatGetter interface
type CpuShares struct{}

// GetStats reads shares metrics from Cpu Group from cpu.shares
func (cpuShares *CpuShares) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	shares, err := parseIntValue(filepath.Join(path, "cpu.shares"))
	if err != nil {
		return err
	}

	stats.Cgroups.CpuStats.CpuShares = shares

	return nil
}
