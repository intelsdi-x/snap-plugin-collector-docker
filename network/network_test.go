/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package network

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockNetworkInterfacesDir = "/tmp/mock_sys_class_net"
	mockProcfsDir            = "/tmp/mock_proc"
)

func TestGetListOfNetworkMetrics(t *testing.T) {

	Convey("List of available network metrics", t, func() {
		networkMetrics := getListOfNetworkMetrics()
		So(len(networkMetrics), ShouldBeGreaterThan, 0)

		Convey("confirm availability of TX metrics", func() {
			So(networkMetrics, ShouldContain, "tx_bytes")
			So(networkMetrics, ShouldContain, "tx_packets")
			So(networkMetrics, ShouldContain, "tx_dropped")
			So(networkMetrics, ShouldContain, "tx_errors")
		})

		Convey("confirm availability of RX metrics", func() {
			So(networkMetrics, ShouldContain, "rx_bytes")
			So(networkMetrics, ShouldContain, "rx_packets")
			So(networkMetrics, ShouldContain, "rx_dropped")
			So(networkMetrics, ShouldContain, "rx_errors")
		})

	})
}

func TestIgnoringDevice(t *testing.T) {

	Convey("Validate devices expected as be ignored", t, func() {
		// device with name started with `lo`, `veth` or `docker` should be ignored
		So(isIgnoredDevice("lo"), ShouldBeTrue)
		So(isIgnoredDevice("docker0"), ShouldBeTrue)
		So(isIgnoredDevice("veth0123456"), ShouldBeTrue)
	})

	Convey("Validate devices expected as NOT be ignored", t, func() {
		So(isIgnoredDevice("eth0"), ShouldBeFalse)
		So(isIgnoredDevice("eno1"), ShouldBeFalse)
		So(isIgnoredDevice("enp2s0"), ShouldBeFalse)
	})
}

func TestListRootNetworkDevices(t *testing.T) {

	Convey("List root network devices", t, func() {

		So(func() { listRootNetworkDevices() }, ShouldNotPanic)

		Convey("return an error when network interface directory is invalid", func() {
			// set invalid networkInterfaceDir (the path does not exist)
			networkInterfacesDir = "/tmp/invalid/path/to/network/devices"
			devs, err := listRootNetworkDevices()
			So(devs, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("return an empty list when there is no device entry", func() {
			networkInterfacesDir = mockNetworkInterfacesDir
			createMockDeviceEntries([]string{})
			defer deleteMockFiles()
			devs, err := listRootNetworkDevices()
			So(err, ShouldBeNil)
			So(devs, ShouldBeEmpty)
		})

		Convey("return not empty list when devices entries are available", func() {
			networkInterfacesDir = mockNetworkInterfacesDir
			mockNetworkDevices := []string{"lo", "docker0", "veth0", "eno1", "eth0", "enp2s0"}
			createMockDeviceEntries(mockNetworkDevices)
			defer deleteMockFiles()
			devs, err := listRootNetworkDevices()
			So(err, ShouldBeNil)
			So(devs, ShouldNotBeEmpty)
			So(len(devs), ShouldEqual, len(mockNetworkDevices))
		})

	})
}

func TestTotalNetworkStats(t *testing.T) {

	Convey("Append `total` to network stats as statistics in total", t, func() {

		Convey("when there is no network interface", func() {
			ifaceStatsInTotal := totalNetworkStats([]wrapper.NetworkInterface{})

			Convey("total stats should be appended", func() {
				So(ifaceStatsInTotal, ShouldNotBeNil)
				So(len(ifaceStatsInTotal), ShouldEqual, 1)
				So(ifaceStatsInTotal[0].Name, ShouldEqual, "total")

				Convey("values of total stats are expected to equal zero", func() {
					So(ifaceStatsInTotal[0].RxBytes, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].RxPackets, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].RxErrors, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].RxDropped, ShouldBeZeroValue)

					So(ifaceStatsInTotal[0].TxBytes, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].TxPackets, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].TxErrors, ShouldBeZeroValue)
					So(ifaceStatsInTotal[0].TxDropped, ShouldBeZeroValue)
				})
			})

		})

		Convey("calulate total stats based on network interface stats", func() {
			// mock network stats per interface
			mockIfaceStats := []wrapper.NetworkInterface{
				wrapper.NetworkInterface{
					Name:      "mockNetInterface1",
					RxBytes:   1,
					RxPackets: 1,
					RxErrors:  1,
					RxDropped: 1,
					TxBytes:   1,
					TxPackets: 1,
					TxErrors:  1,
					TxDropped: 1,
				},

				wrapper.NetworkInterface{
					Name:      "mockNetInterface2",
					RxBytes:   1,
					RxPackets: 1,
					RxErrors:  1,
					RxDropped: 1,
					TxBytes:   1,
					TxPackets: 1,
					TxErrors:  1,
					TxDropped: 1,
				},
			}

			ifaceStatsInTotal := totalNetworkStats(mockIfaceStats)

			Convey("total stats should be appended", func() {
				So(ifaceStatsInTotal, ShouldNotBeNil)
				So(len(ifaceStatsInTotal), ShouldEqual, len(mockIfaceStats)+1)

				for _, ifaceStats := range ifaceStatsInTotal {
					if ifaceStats.Name == "total" {
						Convey("validate values of total stats", func() {
							// there are two mockNetInterfaces with values equal `1` for each metric,
							// so stats in total should equal `2`
							So(ifaceStats.RxBytes, ShouldEqual, 2)
							So(ifaceStats.RxPackets, ShouldEqual, 2)
							So(ifaceStats.RxErrors, ShouldEqual, 2)
							So(ifaceStats.RxDropped, ShouldEqual, 2)

							So(ifaceStats.TxBytes, ShouldEqual, 2)
							So(ifaceStats.TxPackets, ShouldEqual, 2)
							So(ifaceStats.TxErrors, ShouldEqual, 2)
							So(ifaceStats.TxDropped, ShouldEqual, 2)
						})
						continue
					}
				}

			})

		})

	})

}

func TestInterfaceStatsFromDir(t *testing.T) {
	defer deleteMockFiles()
	networkInterfacesDir = mockNetworkInterfacesDir
	mockNetworkDevices := []string{"eno1", "eth0", "enp2s0"}
	mockStatsContent := []byte(`1234`)

	Convey("Get interface stats from networkInterfacesDir", t, func() {

		Convey("create statistics for mock devices", func() {
			err := createMockDeviceStatistics(mockNetworkDevices, mockStatsContent)
			So(err, ShouldBeNil)
		})

		Convey("successful retrieving statistics for available devices", func() {
			for _, device := range mockNetworkDevices {
				stats, err := interfaceStatsFromDir(device)
				So(err, ShouldBeNil)
				So(stats, ShouldNotBeNil)
				So(stats.RxBytes, ShouldEqual, 1234)
				So(stats.TxBytes, ShouldEqual, 1234)
			}
		})

		Convey("return an error when requested device is not available", func() {
			stats, err := interfaceStatsFromDir("invalid_device")
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeNil)
		})

	})
}

func TestNetworkStatsFromRoot(t *testing.T) {
	defer deleteMockFiles()

	networkInterfacesDir = mockNetworkInterfacesDir
	mockNetworkDevices := []string{"eno1", "eth0", "enp2s0"}
	mockNetworkDevicesIgnored := []string{"lo", "docker0", "veth0"}
	mockStatsContent := []byte(`1234`)

	Convey("Get network stats from root", t, func() {

		Convey("create statistics for mock devices", func() {
			err := createMockDeviceStatistics(append(mockNetworkDevices, mockNetworkDevicesIgnored...), mockStatsContent)
			So(err, ShouldBeNil)
		})

		Convey("successful retrieving statistics for available devices", func() {
			stats, err := NetworkStatsFromRoot()
			So(err, ShouldBeNil)
			So(stats, ShouldNotBeEmpty)
			// 4 stats should be returned: for `eno1`, `eth0`, `enp2s0` and `total`
			So(len(stats), ShouldEqual, len(mockNetworkDevices)+1)
		})

		Convey("return an error when there is no available device", func() {
			deleteMockFiles()
			stats, err := NetworkStatsFromRoot()
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeEmpty)
		})

		Convey("return an error when statistics file is not available in device entry path", func() {
			deleteMockFiles()
			createMockDeviceEntries(mockNetworkDevices)
			stats, err := NetworkStatsFromRoot()
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeEmpty)
		})

	})
}

func TestNetworkStatsFromProc(t *testing.T) {
	defer deleteMockFiles()

	// docker container's process ID points to  its network stats in /proc/{pid}/net/dev
	mockPids := []int{1234, 5678, 91011}
	mockDevContent := []byte(`Inter-|   Receive                                                |  Transmit
				face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
				eth0:  424999    4499    0    0    0     0          0         0      648       8    0    0    0     0       0          0
				lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0`)
	mockDevContentLoopback := []byte(`Inter-|   Receive                                                |  Transmit
				face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
				lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0`)

	Convey("Get network stats from root", t, func() {

		Convey("create statistics for mock devices", func() {
			err := createMockProcfsNetDev(mockPids, mockDevContent)
			So(err, ShouldBeNil)
		})

		Convey("successful retrieving statistics for available devices", func() {
			for _, pid := range mockPids {
				stats, err := NetworkStatsFromProc(mockProcfsDir, pid)
				So(err, ShouldBeNil)
				So(stats, ShouldNotBeEmpty)
				// stats should be returned: for `eth0` and `total`; `lo` should be ignored
				So(len(stats), ShouldEqual, 2)
			}

		})

		Convey("return an error when the given PID does not exist", func() {
			stats, err := NetworkStatsFromProc(mockProcfsDir, 0)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeEmpty)
		})

		Convey("return an error when no network interface found", func() {
			mockPid := 1
			Convey("create net/dev which contains only loopback", func() {
				err := createMockProcfsNetDev([]int{mockPid}, mockDevContentLoopback)
				So(err, ShouldBeNil)
			})
			stats, err := NetworkStatsFromProc(mockProcfsDir, mockPid)
			So(err, ShouldNotBeNil)
			So(stats, ShouldBeEmpty)
			So(err.Error(), ShouldEqual, "No network interface found")
		})
	})
}

// createMockDeviceStatistics creates for the given devices' names statistics file with given content
// under the following path: /mockNetworkInterfacesDir/{device}/statistics
func createMockDeviceStatistics(devices []string, content []byte) error {
	deleteMockFiles()
	for _, device := range devices {
		pathToDeviceStats := filepath.Join(mockNetworkInterfacesDir, device, "statistics")
		if err := os.MkdirAll(pathToDeviceStats, os.ModePerm); err != nil {
			return err
		}

		for _, statName := range networkMetrics {
			if err := createFile(pathToDeviceStats, statName, content); err != nil {
				return err
			}
		}
	}

	return nil
}

// createMockProcfsNetDev creates for the given process IDs net/dev statistics with given content
// under the following path: /mockProcfsDir/{pid}/net/dev
func createMockProcfsNetDev(pids []int, content []byte) error {
	deleteMockFiles()
	for _, pid := range pids {
		pathToProcessNetDev := filepath.Join(mockProcfsDir, "proc", fmt.Sprintf("%d", pid), "net")
		if err := os.MkdirAll(pathToProcessNetDev, os.ModePerm); err != nil {
			return err
		}

		if err := createFile(pathToProcessNetDev, "dev", content); err != nil {
			return err
		}
	}

	return nil
}

// createMockDeviceEntries creates folder named as device for the given devices
// under the following path: /mockNetworkInterfacesDir/{device}
func createMockDeviceEntries(devices []string) error {
	deleteMockFiles()

	if err := os.MkdirAll(mockNetworkInterfacesDir, os.ModePerm); err != nil {
		return err
	}

	for _, device := range devices {
		devEntry := filepath.Join(mockNetworkInterfacesDir, device)
		if err := os.Mkdir(devEntry, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// createMockProcfsNetTCP creates for the given process IDs net/tcp and net/tcp6 statistics with given content
// under the following path: /mockProcfsDir/{pid}/net/tcp and /mockProcfsDir/{pid}/net/tcp6
func createMockProcfsNetTCP(pids []int, content []byte) error {
	deleteMockFiles()
	for _, pid := range pids {
		pathToProcessNetDev := filepath.Join(mockProcfsDir, "proc", fmt.Sprintf("%d", pid), "net")
		if err := os.MkdirAll(pathToProcessNetDev, os.ModePerm); err != nil {
			return err
		}

		// create tcp file
		if err := createFile(pathToProcessNetDev, "tcp", content); err != nil {
			return err
		}

		// create tcp6 file
		if err := createFile(pathToProcessNetDev, "tcp6", content); err != nil {
			return err
		}
	}

	return nil
}

// deleteMockFiles removes mock files
func deleteMockFiles() {
	os.RemoveAll(mockNetworkInterfacesDir)
	os.RemoveAll(mockProcfsDir)
}

// createFile creates file and writes to it a given content
func createFile(path string, name string, content []byte) error {
	// create file in a given path
	f, err := os.Create(filepath.Join(path, name))
	if err == nil {
		// when file was created successfully, write a content to it
		_, err = f.Write(content)
	}

	return err
}
