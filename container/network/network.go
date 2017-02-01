/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation
Copyright 2012-2013 Rackspace, Inc.

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

// Package contains code from Google Cadvisor (https://github.com/google/cadvisor) with following:
// - functions collecting network statistics

// Package network provides network statistics
package network

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
	utils "github.com/intelsdi-x/snap-plugin-utilities/ns"

	log "github.com/Sirupsen/logrus"
)

const (
	// expected format of the line of net stats file
	numberOfFields = 17
	// expected 4 type of stats: (Rx/Tx) packets, bytes, dropped, errors
	numberOfStatsType = 4
	indexOfRxStats    = 1
	indexOfTxStats    = 9
)

var (
	// networkInterfaceDir points to network devices and its stats (declaring as var for mock purpose)
	networkInterfacesDir = "/sys/class/net"

	// networkMetrics is a list of available network metrics (rx_bytes, tx_bytes, etc.)
	networkMetrics = getListOfNetworkMetrics()
)

func getListOfNetworkMetrics() []string {
	metrics := []string{}
	utils.FromCompositionTags(container.NetworkInterface{}, "", &metrics)
	return metrics
}

type Network struct{}

func (n *Network) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	pid, err := opts.GetIntValue("pid")
	if err != nil {
		return err
	}

	isHost, err := opts.GetBoolValue("is_host")
	if err != nil {
		return err
	}

	procfs, err := opts.GetStringValue("procfs")
	if err != nil {
		return err
	}

	if !isHost {
		path := filepath.Join(procfs, strconv.Itoa(pid))
		stats.Network, err = NetworkStatsFromProc(path)
		if err != nil {
			// only log error message
			log.WithFields(log.Fields{
				"module": "network",
				"block":  "GetStats",
			}).Errorf("Unable to get network stats, pid %d: %s", pid, err)
		}

	} else {
		stats.Network, err = NetworkStatsFromRoot()
		if err != nil {
			// only log error message
			log.WithFields(log.Fields{
				"module": "network",
				"block":  "GetStats",
			}).Errorf("Unable to get network stats for host: %s", err)
		}
	}

	return nil
}

// NetworkStatsFromProc returns network statistics (e.g. tx_bytes, rx_bytes, etc.) per each interface and aggregated in total
// for a given path combined from given rootFs and pid of docker process as `<rootFs>/<set_procfs_mountpoint>/<pid>/net/dev`
func NetworkStatsFromProc(path string) ([]container.NetworkInterface, error) {
	netStatsFile := filepath.Join(path, "/net/dev")
	ifaceStats, err := scanInterfaceStats(netStatsFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read network stats: %v", err)
	}

	if len(ifaceStats) == 0 {
		return nil, errors.New("No network interface found")
	}

	return totalNetworkStats(ifaceStats), nil
}

// NetworkStatsFromRoot returns network statistics (e.g. tx_bytes, rx_bytes, etc.) per each interface and aggregated in total
// for root (a docker host)
func NetworkStatsFromRoot() (ifaceStats []container.NetworkInterface, _ error) {
	devNames, err := listRootNetworkDevices()
	if err != nil {
		return nil, err
	}
	ifaceStats = []container.NetworkInterface{}
	for _, name := range devNames {
		if isIgnoredDevice(name) {
			continue
		}
		if stats, err := interfaceStatsFromDir(name); err != nil {
			return nil, err
		} else {
			ifaceStats = append(ifaceStats, *stats)
		}
	}
	return totalNetworkStats(ifaceStats), nil
}

// totalNetworkStats calculates summary of network stats (sum over all net interfaces) and returns
func totalNetworkStats(ifaceStats []container.NetworkInterface) (ifaceStatsInTotal []container.NetworkInterface) {
	total := container.NetworkInterface{
		Name: "total",
	}

	for _, iface := range ifaceStats {
		total.RxBytes += iface.RxBytes
		total.RxPackets += iface.RxPackets
		total.RxDropped += iface.RxDropped
		total.RxErrors += iface.RxErrors
		total.TxBytes += iface.TxBytes
		total.TxPackets += iface.TxPackets
		total.TxDropped += iface.TxDropped
		total.TxErrors += iface.TxErrors
	}

	return append(ifaceStats, total)
}

func listRootNetworkDevices() (devNames []string, _ error) {
	entries, err := ioutil.ReadDir(networkInterfacesDir)
	if err != nil {
		return nil, err
	}
	devNames = []string{}
	for _, e := range entries {
		if e.Mode()&os.ModeSymlink == os.ModeSymlink {
			e, err = os.Stat(filepath.Join(networkInterfacesDir, e.Name()))
			if err != nil || !e.IsDir() {
				continue
			}
			devNames = append(devNames, e.Name())
		} else if e.IsDir() {
			devNames = append(devNames, e.Name())
		}
	}
	return devNames, nil
}

func interfaceStatsFromDir(ifaceName string) (*container.NetworkInterface, error) {
	stats := container.NetworkInterface{Name: ifaceName}
	statsValues := map[string]uint64{}
	for _, metric := range networkMetrics {
		if metric == "name" {
			continue
		}
		val, err := readUintFromFile(filepath.Join(networkInterfacesDir, ifaceName, "statistics", metric), 64)
		if err != nil {
			return nil, fmt.Errorf("couldn't read interface statistics %s/%s: %v", ifaceName, metric, err)
		}
		statsValues[metric] = val
	}
	setIfaceStatsFromMap(&stats, statsValues)
	return &stats, nil
}

func setIfaceStatsFromMap(stats *container.NetworkInterface, values map[string]uint64) {
	stats.RxBytes = values["rx_bytes"]
	stats.RxErrors = values["rx_errors"]
	stats.RxPackets = values["rx_packets"]
	stats.RxDropped = values["rx_dropped"]
	stats.TxBytes = values["tx_bytes"]
	stats.TxErrors = values["tx_errors"]
	stats.TxPackets = values["tx_packets"]
	stats.TxDropped = values["tx_dropped"]
}

func isIgnoredDevice(ifName string) bool {
	ignoredDevicePrefixes := []string{"lo", "veth", "docker"}
	for _, prefix := range ignoredDevicePrefixes {
		if strings.HasPrefix(strings.ToLower(ifName), prefix) {
			return true
		}
	}
	return false
}

func scanInterfaceStats(netStatsFile string) ([]container.NetworkInterface, error) {
	file, err := os.Open(netStatsFile)
	if err != nil {
		return nil, fmt.Errorf("failure opening %s: %v", netStatsFile, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Discard header lines
	for i := 0; i < 2; i++ {
		if b := scanner.Scan(); !b {
			return nil, scanner.Err()
		}
	}

	stats := []container.NetworkInterface{}
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, ":", "", -1)

		fields := strings.Fields(line)
		// If the format of the  line is invalid then don't trust any of the stats
		// in this file.
		if len(fields) != numberOfFields {
			return nil, fmt.Errorf("invalid interface stats line: %v", line)
		}

		devName := fields[0]

		if isIgnoredDevice(devName) {
			continue
		}

		i := container.NetworkInterface{
			Name: devName,
		}

		// take fields [1:5] for rx stats and [9:13] for tx stats
		rxStatsFields := fields[indexOfRxStats : indexOfRxStats+numberOfStatsType]
		txStatsFields := fields[indexOfTxStats : indexOfTxStats+numberOfStatsType]

		statFields := append(rxStatsFields, txStatsFields...)
		statPointers := []*uint64{
			&i.RxBytes, &i.RxPackets, &i.RxErrors, &i.RxDropped,
			&i.TxBytes, &i.TxPackets, &i.TxErrors, &i.TxDropped,
		}

		err := setInterfaceStatValues(statFields, statPointers)
		if err != nil {
			return nil, fmt.Errorf("cannot parse interface stats (%v): %v", err, line)
		}

		stats = append(stats, i)
	}

	return stats, nil
}

func setInterfaceStatValues(fields []string, pointers []*uint64) error {
	for i, v := range fields {
		val, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		*pointers[i] = val
	}
	return nil
}

func readUintFromFile(path string, bits int) (uint64, error) {
	valb, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(valb)), 10, bits)
}
