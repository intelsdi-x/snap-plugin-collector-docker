// +build linux

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

package collector

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	utils "github.com/intelsdi-x/snap-plugin-utilities/ns"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
	"github.com/intelsdi-x/snap-plugin-collector-docker/container/cgroupfs"
	"github.com/intelsdi-x/snap-plugin-collector-docker/container/fs"
	"github.com/intelsdi-x/snap-plugin-collector-docker/container/network"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

const (
	// namespace vendor prefix
	PLUGIN_VENDOR = "intel"
	// namespace plugin name
	PLUGIN_NAME = "docker"
	// version of plugin
	PLUGIN_VERSION = 7

	// each metric starts with prefix "/intel/docker/<docker_id>"
	lengthOfNsPrefix = 3
)

var getters map[string]container.StatGetter = map[string]container.StatGetter{
	"throttling_data": &cgroupfs.Cpu{},
	"cpu_usage":       &cgroupfs.CpuAcct{},
	"cpu_shares":      &cgroupfs.CpuShares{},
	"cache":           &cgroupfs.MemoryCache{},
	"usage":           &cgroupfs.MemoryUsage{},
	"swap_usage":      &cgroupfs.SwapMemUsage{},
	"kernel_usage":    &cgroupfs.KernelMemUsage{},
	"statistics":      &cgroupfs.Memory{},
	"blkio_stats":     &cgroupfs.Blkio{},
	"hugetlb_stats":   &cgroupfs.HugeTlb{},
	"pids_stats":      &cgroupfs.Pids{},
	"cpuset_stats":    &cgroupfs.CpuSet{},
	"network":         &network.Network{},
	"tcp":             &network.Tcp{StatsFile: "net/tcp"},
	"tcp6":            &network.Tcp{StatsFile: "net/tcp6"},
	"filesystem":      &fs.DiskUsageCollector{},
}

var names map[string]string = map[string]string{
	"throttling_data": "cpu",
	"cpu_usage":       "cpuacct",
	"cpu_shares":      "cpu",
	"cache":           "memory",
	"usage":           "memory",
	"swap_usage":      "memory",
	"kernel_usage":    "memory",
	"statistics":      "memory",
	"blkio_stats":     "blkio",
	"hugetlb_stats":   "hugetlb",
	"pids_stats":      "pids",
	"cpuset_stats":    "cpuset",
	"spec":            "spec",
	"network":         "network",
	"tcp":             "tcp",
	"tcp6":            "tcp6",
	"filesystem":      "filesystem",
}

// New returns initialized docker plugin
func New() plugin.Collector {
	return &collector{
		containers: map[string]*container.ContainerData{},
		mounts:     map[string]string{},
	}
}

// CollectMetrics retrieves values of requested metrics
func (c *collector) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	var err error
	metrics := []plugin.Metric{}
	// setup docker client based on config only once
	if c.client == nil {
		c.conf, err = getDockerConfig(mts[0].Config)
		if err != nil {
			log.WithFields(log.Fields{
				"block":    "CollectMetrics",
				"function": "getDockerConfig",
			}).Error(err)
			return nil, err
		}
		err = initClient(c, c.conf["endpoint"])
		if err != nil {
			log.WithFields(log.Fields{
				"block":    "CollectMetrics",
				"function": "initClient",
			}).Error(err)
			return nil, err
		}
	}

	// get list of all running containers
	c.containers, err = c.client.ListContainersAsMap()
	if err != nil {
		log.WithFields(log.Fields{
			"block":    "CollectMetrics",
			"function": "ListContainersAsMap",
		}).Error(err)
		return nil, err
	}
	// group requested metrics by docker id
	ridGroup, err := c.getRidGroup(mts...)
	if err != nil {
		log.WithFields(log.Fields{
			"block":    "CollectMetrics",
			"function": "getRidGroup",
		}).Error(err)
		return nil, err
	}

	// collect requested metrics per docker id
	err = c.collect(ridGroup, c.conf["procfs"])
	if err != nil {
		log.WithFields(log.Fields{
			"block":    "CollectMetrics",
			"function": "collect",
		}).Error(err)
		return nil, err
	}

	for _, mt := range mts {
		ridGroup, err := c.getRidGroup(mt)
		if err != nil {
			log.WithFields(log.Fields{
				"block":    "CollectMetrics",
				"function": "getRidGroup",
			}).Error(err)
			return nil, err
		}

		for rid := range ridGroup {
			ns := make([]plugin.NamespaceElement, len(mt.Namespace))
			copy(ns, mt.Namespace)
			ns[2].Value = rid

			// omit "spec" metrics for root
			if rid == "root" && mt.Namespace[lengthOfNsPrefix].Value == "spec" {
				continue
			}

			// omit "pids stats" for host
			if rid == "root" && mt.Namespace[lengthOfNsPrefix].Value == "pids_stats" {
				log.WithFields(log.Fields{
					"block": "CollectMetrics",
				}).Warnf("pids stats are not avaialble for host")
				continue
			}

			isDynamic, indexes := mt.Namespace[lengthOfNsPrefix:].IsDynamic()

			metricName := mt.Namespace.Strings()[lengthOfNsPrefix:]

			// remove added static element (`value`)
			if metricName[len(metricName)-1] == "value" {
				metricName = metricName[:len(metricName)-1]
			}

			if !isDynamic {
				metric := plugin.Metric{
					Timestamp: time.Now(),
					Namespace: ns,
					Data:      utils.GetValueByNamespace(c.containers[rid], metricName),
					Tags:      mt.Tags,
					Config:    mt.Config,
					Version:   PLUGIN_VERSION,
				}

				metrics = append(metrics, metric)
				continue

			}

			// take the element of metricName which precedes the first dynamic element
			// e.g. {"filesystem", "*", "usage"}
			// 	-> statsType will be "filesystem",
			// 	-> scope of metricName will be decreased to {"*", "usage"}
			indexOfDynamicElement := indexes[0]
			statsType := metricName[indexOfDynamicElement-1]
			metricName = metricName[indexOfDynamicElement:]

			switch statsType {
			case "filesystem":
				// get docker filesystem statistics
				devices := []string{}

				if metricName[0] == "*" {
					// when device name is requested as as asterisk - take all available filesystem devices
					for deviceName := range c.containers[rid].Stats.Filesystem {
						devices = append(devices, deviceName)
					}
				} else {
					// device name is requested explicitly
					device := metricName[0]
					fs_device := c.containers[rid].Stats.Filesystem[device]
					if fs_device.Device == "" {
						return nil, fmt.Errorf("In metric %s the given device name is invalid (no stats for this device)", strings.Join(mt.Namespace.Strings(), "/"))
					}

					devices = append(devices, metricName[0])
				}

				for _, device := range devices {
					rns := make([]plugin.NamespaceElement, len(ns))
					copy(rns, ns)

					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = device

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: rns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Filesystem[device], metricName[1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "labels":
				// get docker labels
				labelKeys := []string{}
				if metricName[0] == "*" {
					// when label key is requested as an asterisk - take all available labels
					for labelKey := range c.containers[rid].Specification.Labels {
						labelKeys = append(labelKeys, labelKey)
					}
				} else {
					labelKey := metricName[0]
					c_label := c.containers[rid].Specification.Labels[labelKey]
					if c_label == "" {
						return nil, fmt.Errorf("In metric %s the given label is invalid (no value for this label key)", strings.Join(mt.Namespace.Strings(), "/"))
					}

					labelKeys = append(labelKeys, metricName[0])
				}

				for _, labelKey := range labelKeys {
					rns := make([]plugin.NamespaceElement, len(ns))
					copy(rns, ns)
					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = utils.ReplaceNotAllowedCharsInNamespacePart(labelKey)
					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: rns,
						Data:      c.containers[rid].Specification.Labels[labelKey],
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}

					metrics = append(metrics, metric)
				}

			case "network":
				//get docker network tx/rx statistics
				netInterfaces := []string{}
				ifaceMap := map[string]container.NetworkInterface{}
				for _, iface := range c.containers[rid].Stats.Network {
					ifaceMap[iface.Name] = iface
				}

				// support wildcard on interface name
				if metricName[0] == "*" {
					for _, netInterface := range c.containers[rid].Stats.Network {
						netInterfaces = append(netInterfaces, netInterface.Name)
					}
				} else {
					netInterface := metricName[0]
					if _, ok := ifaceMap[netInterface]; !ok {
						return nil, fmt.Errorf("In metric %s the given network interface is invalid (no stats for this net interface)", strings.Join(mt.Namespace.Strings(), "/"))
					}
					netInterfaces = append(netInterfaces, metricName[0])
				}

				for _, ifaceName := range netInterfaces {
					rns := make([]plugin.NamespaceElement, len(ns))
					copy(rns, ns)
					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = ifaceName
					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: rns,
						Data:      utils.GetValueByNamespace(ifaceMap[ifaceName], metricName[1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "per_cpu":
				numOfCPUs := len(c.containers[rid].Stats.Cgroups.CpuStats.CpuUsage.PerCpu) - 1
				if metricName[0] == "*" {
					// when cpu ID is requested as an asterisk - take all available
					for cpuID, val := range c.containers[rid].Stats.Cgroups.CpuStats.CpuUsage.PerCpu {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(cpuID)

						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      val,
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					cpuID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given cpu id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if cpuID > numOfCPUs || cpuID < 0 {
						return nil, fmt.Errorf("In metric %s the given cpu id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfCPUs)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      c.containers[rid].Stats.Cgroups.CpuStats.CpuUsage.PerCpu[cpuID],
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "hugetlb_stats":
				sizes := []string{}
				if metricName[0] == "*" {
					for size := range c.containers[rid].Stats.Cgroups.HugetlbStats {
						sizes = append(sizes, size)
					}
				} else {
					size := metricName[0]
					if _, ok := c.containers[rid].Stats.Cgroups.HugetlbStats[size]; !ok {
						return nil, fmt.Errorf("In metric %s the given hugetlb size is invalid (no stats for this size)", strings.Join(mt.Namespace.Strings(), "/"))
					}
					sizes = append(sizes, size)
				}

				for _, size := range sizes {
					rns := make([]plugin.NamespaceElement, len(ns))
					copy(rns, ns)
					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = size
					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: rns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.HugetlbStats[size], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "io_merged_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoMergedRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, imr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoMergedRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(imr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoMergedRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_service_bytes_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceBytesRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, isbr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceBytesRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(isbr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceBytesRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_serviced_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoServicedRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, isr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoServicedRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(isr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoServicedRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_queue_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoQueuedRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, iqr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoQueuedRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(iqr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoQueuedRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_service_time_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceTimeRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, istr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceTimeRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(istr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoServiceTimeRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_wait_time_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoWaitTimeRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, iwtr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoWaitTimeRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(iwtr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoWaitTimeRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "io_time_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.IoTimeRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, itr := range c.containers[rid].Stats.Cgroups.BlkioStats.IoTimeRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(itr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.IoTimeRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			case "sectors_recursive":
				numOfDevices := len(c.containers[rid].Stats.Cgroups.BlkioStats.SectorsRecursive) - 1
				if metricName[0] == "*" {
					for deviceID, sr := range c.containers[rid].Stats.Cgroups.BlkioStats.SectorsRecursive {
						rns := make([]plugin.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(deviceID)
						metric := plugin.Metric{
							Timestamp: time.Now(),
							Namespace: rns,
							Data:      utils.GetValueByNamespace(sr, mt.Namespace.Strings()[len(ns)-1:]),
							Tags:      mt.Tags,
							Config:    mt.Config,
							Version:   PLUGIN_VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					deviceID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, err=%v", strings.Join(mt.Namespace.Strings(), "/"), err)
					}
					if deviceID > numOfDevices || deviceID < 0 {
						return nil, fmt.Errorf("In metric %s the given device id is invalid, expected value in range 0-%d", strings.Join(mt.Namespace.Strings(), "/"), numOfDevices)
					}

					metric := plugin.Metric{
						Timestamp: time.Now(),
						Namespace: ns,
						Data:      utils.GetValueByNamespace(c.containers[rid].Stats.Cgroups.BlkioStats.SectorsRecursive[deviceID], mt.Namespace.Strings()[len(ns)-1:]),
						Tags:      mt.Tags,
						Config:    mt.Config,
						Version:   PLUGIN_VERSION,
					}
					metrics = append(metrics, metric)
				}
			} // the end of switch statsType
		} // the end of range over ids
	}

	if len(metrics) == 0 {
		return nil, fmt.Errorf("No metrics found")
	}

	return metrics, nil
}

// GetMetricTypes returns list of available metrics
func (c *collector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	var err error
	var metricTypes []plugin.Metric

	// initialize containerData struct
	data := container.ContainerData{
		Stats: container.NewStatistics(),
	}

	dockerMetrics := []string{}
	utils.FromCompositeObject(data, "", &dockerMetrics)
	nscreator := nsCreator{dynamicElements: definedDynamicElements}
	for _, metricName := range dockerMetrics {
		ns := plugin.NewNamespace(PLUGIN_VENDOR, PLUGIN_NAME).
			AddDynamicElement("docker_id", "an id of docker container")

		if ns, err = nscreator.createMetricNamespace(ns, metricName); err != nil {
			// skip this metric name which is not supported
			log.WithFields(log.Fields{
				"block": "GetMetricTypes",
			}).Warnf("Error in creating metric %s: err=%s", metricName, err)
			continue
		}
		metricType := plugin.Metric{
			Namespace: ns,
			Version:   PLUGIN_VERSION,
		}
		metricTypes = append(metricTypes, metricType)
	}

	return metricTypes, nil
}

// GetConfigPolicy returns plugin config policy
func (c *collector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	configKey := []string{"intel", "docker"}

	policy.AddNewStringRule(configKey,
		"endpoint",
		false,
		plugin.SetDefaultString("unix:///var/run/docker.sock"))

	policy.AddNewStringRule(configKey,
		"procfs",
		false,
		plugin.SetDefaultString("/proc"))

	return *policy, nil
}

type collector struct {
	containers map[string]*container.ContainerData // holds data for a container under its short id
	client     container.DockerClientInterface     // client for communication with docker (basic info, stats, mount points)
	cgroupfs   string                              // CgroupDriver from docker engine
	driver     string                              // Driver from docker engine
	rootDir    string                              // Storage mount point for docker containers
	mounts     map[string]string                   // cache for cgroup mountpoints
	conf       map[string]string                   // plugin configuration passed with metrics
}

// getRidGroup returns quested metrics grouped by docker ids
func (c *collector) getRidGroup(mt ...plugin.Metric) (map[string]map[string]struct{}, error) {
	ridGroup := make(map[string]map[string]struct{})
	for _, m := range mt {
		ns := m.Namespace.Strings()
		if len(ns) < lengthOfNsPrefix+2 {
			return nil, fmt.Errorf("Invalid name of metric %+s", strings.Join(ns, "/"))
		}

		rid := ns[2]

		group, err := getQueryGroup(ns[3:])
		if err != nil {
			return nil, err
		}

		switch rid {
		case "*":
			for id := range c.containers {
				appendIfMissing(ridGroup, id, group)
			}
		case "root":
			appendIfMissing(ridGroup, "root", group)
		default:
			shortID, err := container.GetShortID(rid)
			if err != nil {
				return nil, err
			}

			if _, exist := c.containers[shortID]; !exist {
				return nil, fmt.Errorf("Docker container %+s cannot be found", rid)
			}

			appendIfMissing(ridGroup, shortID, group)
		}
	}

	if len(ridGroup) == 0 {
		return nil, fmt.Errorf("can't retrieve docker ids and requestd metrics")
	}

	return ridGroup, nil
}

func (c *collector) collect(ridGroup map[string]map[string]struct{}, procfs string) error {
	var err error
	var cont *docker.Container
	for rid, groups := range ridGroup {
		opts := make(container.GetStatOpt)
		opts["procfs"] = procfs
		opts["root_dir"] = c.rootDir

		if rid == "root" {
			opts["is_host"] = true
			opts["pid"] = -1
			opts["container_id"] = "root"
			opts["container_drv"] = c.driver
		} else {
			cont, err = c.client.InspectContainer(rid)
			if err != nil {
				return err
			}
			opts["is_host"] = false
			opts["pid"] = cont.State.Pid
			opts["container_id"] = cont.ID
			opts["container_drv"] = cont.Driver
		}

		for group := range groups {
			// during initialization of docker client information about running containers is collected
			if group == "spec" {
				continue
			}

			if group == "pids_stats" && rid == "root" {
				continue
			}

			if group != "network" && group != "tcp" && group != "tcp6" && group != "filesystem" {
				cgroup := names[group]
				// try to find cgroup mount point in cache
				cpath, exists := c.mounts[cgroup]
				if !exists {
					cpath, err = c.client.FindCgroupMountpoint(procfs, cgroup)
					if err != nil {
						return err
					}
					c.mounts[cgroup] = cpath
				}

				if rid != "root" {
					cpath, err = c.client.FindControllerMountpoint(cgroup, strconv.Itoa(cont.State.Pid), procfs)
					if err != nil {
						return err
					}
				}
				opts["cgroup_path"] = cpath
			}

			shortID, err := container.GetShortID(rid)
			if err != nil {
				return err
			}

			err = getters[group].GetStats(c.containers[shortID].Stats, opts)
			// only log error when it was not possible to access metric source
			if err != nil {
				log.WithFields(log.Fields{
					"block": "collect",
				}).Error(err)
			}
		}
	}

	return nil
}
