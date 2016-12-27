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
	"os"
	"path/filepath"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

// Memory implements StatGetter interface
type Memory struct{}

// GetStats reads general memory metrics from Memory Group from memory.stat
func (mem *Memory) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(path, "memory.stat"))
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

		stats.Cgroups.MemoryStats.Stats[param] = value
	}

	// calculate additional stats memory:working_set based on memory_stats
	var workingSet uint64
	if totalInactiveAnon, ok := stats.Cgroups.MemoryStats.Stats["total_inactive_anon"]; ok {
		workingSet = stats.Cgroups.MemoryStats.Usage.Usage
		if workingSet < totalInactiveAnon {
			workingSet = 0
		} else {
			workingSet -= totalInactiveAnon
		}

		if totalInactiveFile, ok := stats.Cgroups.MemoryStats.Stats["total_inactive_file"]; ok {
			if workingSet < totalInactiveFile {
				workingSet = 0
			} else {
				workingSet -= totalInactiveFile
			}
		}
	}
	stats.Cgroups.MemoryStats.Stats["working_set"] = workingSet

	return nil
}

// MemoryCache implements StatGetter interface
type MemoryCache struct{}

// GetStats reads memory cache metric from Memory Group from memory.stat
func (memCa *MemoryCache) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}
	// if memory.stat where already collected, check for cache in map
	if stats.Cgroups.MemoryStats.Stats["cache"] != 0 {
		stats.Cgroups.MemoryStats.Cache = stats.Cgroups.MemoryStats.Stats["cache"]
		return nil
	}

	f, err := os.Open(filepath.Join(path, "memory.stat"))
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "cache") {
			_, val, err := parseEntry(line)
			if err != nil {
				return err
			}
			stats.Cgroups.MemoryStats.Cache = val
			break
		}
	}

	return nil
}

// MemoryUsage implements StatGetter interface
type MemoryUsage struct{}

// GetStats reads memory usage metrics from Memory Group from memory.usage_in_bytes, memory.failcnt, memory.max_usage_in_bytes
func (memu *MemoryUsage) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	memoryData, err := getMemoryData(path, "")
	if err != nil {
		return err
	}
	stats.Cgroups.MemoryStats.Usage = memoryData

	return nil
}

// SwapMemUsage implements StatGetter interface
type SwapMemUsage struct{}

// GetStats reads memory swap usage metrics from Memory Group from memory.memsw.usage_in_bytes, memory.memsw.failcnt, memory.memsw.max_usage_in_bytes
func (memu *SwapMemUsage) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	memoryData, err := getMemoryData(path, "memsw")
	if err != nil {
		return err
	}
	stats.Cgroups.MemoryStats.SwapUsage = memoryData

	return nil
}

// KernelMemUsage implements StatGetter interface
type KernelMemUsage struct{}

// GetStats reads memory kernel usage metrics from Memory Group from memory.kmem.usage_in_bytes, memory.kmem.failcnt, memory.kmem.max_usage_in_bytes
func (memu *KernelMemUsage) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	path, err := opts.GetStringValue("cgroup_path")
	if err != nil {
		return err
	}

	memoryData, err := getMemoryData(path, "kmem")
	if err != nil {
		return err
	}
	stats.Cgroups.MemoryStats.KernelUsage = memoryData

	return nil
}

func getMemoryData(path, name string) (container.MemoryData, error) {
	moduleName := "memory"
	if name != "" {
		moduleName = strings.Join([]string{"memory", name}, ".")
	}

	memoryData := container.MemoryData{}

	usage, err := parseIntValue(filepath.Join(path, strings.Join([]string{moduleName, "usage_in_bytes"}, ".")))
	if err != nil {
		return memoryData, err
	}

	maxUsage, err := parseIntValue(filepath.Join(path, strings.Join([]string{moduleName, "max_usage_in_bytes"}, ".")))
	if err != nil {
		return memoryData, err
	}

	failcnt, err := parseIntValue(filepath.Join(path, strings.Join([]string{moduleName, "failcnt"}, ".")))
	if err != nil {
		return memoryData, err
	}

	memoryData.Usage = usage
	memoryData.MaxUsage = maxUsage
	memoryData.Failcnt = failcnt

	return memoryData, nil
}
