// +build unit

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

package docker

import (
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/opencontainers/runc/libcontainer/cgroups"

	"github.com/intelsdi-x/pulse/control/plugin"

	. "github.com/intelsdi-x/pulse-plugin-collector-docker/client"
	. "github.com/intelsdi-x/pulse-plugin-collector-docker/wrapper"
	. "github.com/intelsdi-x/pulse-plugin-collector-docker/tools"
	. "github.com/intelsdi-x/pulse-plugin-collector-docker/mocks"



)

func TestExtendDockerIdProper(t *testing.T){

	Convey("Given short docker id", t, func(){

		shortId := "1234567890ab"

		Convey("and containers info with proper extension", func() {

			proper := "1234567890ab9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"
			other := "31068893a2bc9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"

			ci := []ContainerInfo{
					ContainerInfo{Id: other},
					ContainerInfo{Id: proper},
			}

			Convey("When docker id is extended", func() {
				d := docker{containersInfo: ci}
				longId, err := d.extendDockerId(shortId)

				Convey("Then error should not be reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then proper long id is returned", func() {
					So(longId, ShouldEqual, proper)
				})
			})
		})
	})
}

func TestExtendDockerIdWrong(t *testing.T){

	Convey("Given incorrect short docker id", t, func(){

		wrongShortId := "wrongid12334"

		Convey("and containers info with proper extension", func() {

			proper := "1234567890ab9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"
			other := "31068893a2bc9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"

			ci := []ContainerInfo{
				ContainerInfo{Id: other},
				ContainerInfo{Id: proper},
			}

			Convey("When docker id is extended", func() {
				d := docker{containersInfo: ci}
				longId, err := d.extendDockerId(wrongShortId)

				Convey("Then error should be reported", func() {
					So(err, ShouldNotBeNil)
				})

				Convey("Then returned value is empty", func() {
					So(longId, ShouldBeEmpty)
				})
			})
		})
	})
}

func TestGetStats(t *testing.T){

	Convey("Given docker id, stats, client", t, func() {

		dockerId := "1234567890ab"
		mountPoint := "mount/point/path"
		mockStats := new(StatsMock)
		mockClient := new(ClientMock)
		stats := cgroups.NewStats()

		mockClient.On("FindCgroupMountpoint", "cpu").Return(mountPoint, nil)
		mockStats.On("GetStats", mock.AnythingOfType("string"), stats).Return(nil).Run(
			func(args mock.Arguments) {
				arg := args.Get(1).(*cgroups.Stats)
				arg.CpuStats.CpuUsage.TotalUsage = 43
				arg.CpuStats.CpuUsage.PercpuUsage = []uint64{99, 88}
			})

		Convey("and cgroups stats wrapper", func() {

			mockWrapper := map[string]Stats{"cpu": mockStats}

			Convey("When docker stats are requested", func() {

				d := &docker{
					stats:      		stats,
					client:            	mockClient,
					tools:				new(MyTools),
					containersInfo: 	[]ContainerInfo{},
					groupWrap:          mockWrapper,
					hostname:           "",
				}

				err := d.getStats(dockerId)

				Convey("Then no error should be reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then stats should be set", func() {
					So(stats.CpuStats.CpuUsage.TotalUsage, ShouldEqual, 43)
					So(stats.CpuStats.CpuUsage.PercpuUsage[0], ShouldEqual, 99)
					So(stats.CpuStats.CpuUsage.PercpuUsage[1], ShouldEqual, 88)
				})
			})
		})
	})
}

func TestCollectMetrics(t *testing.T){

	Convey("Given 1234567890ab/cpu_stats/cpu_usage/total_usage metric type", t, func() {

		mountPoint := "cgroup/mount/point/path"
		shortDockerId := "1234567890ab"
		longDockerId := "1234567890ab9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"
		ns := []string{NS_VENDOR, NS_CLASS, NS_PLUGIN, shortDockerId, "cpu_stats", "cpu_usage", "total_usage"}
		metricTypes := []plugin.PluginMetricType{plugin.PluginMetricType{Namespace_: ns}}

		Convey("and docker plugin intitialized", func() {

			stats := cgroups.NewStats()

			mockClient := new(ClientMock)
			mockStats := new(StatsMock)
			mockTools := new(ToolsMock)
			mockWrapper := map[string]Stats{"cpu": mockStats}

			mockClient.On("FindCgroupMountpoint", "cpu").Return(mountPoint, nil)

			mockStats.On("GetStats", mock.AnythingOfType("string"), stats).Return(nil)

			// TODO - sprawdzic jak mozna by wykorzystac
			mockTools.On("GetValueByNamespace", mock.AnythingOfType("*cgroups.Stats"), mock.Anything).Return(43)

			d := &docker{
				stats: 				stats,
				client: 			mockClient,
				tools:				mockTools,
				containersInfo: 	[]ContainerInfo{ContainerInfo{Id: longDockerId}},
				groupWrap: 			mockWrapper,
				hostname: 			"",
			}

			Convey("When CollectMetric is called", func() {
				mts, err := d.CollectMetrics(metricTypes)

				Convey("Then error should not be reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("One metric should be returned", func() {
					So(len(mts),ShouldEqual, 1)
				})

				Convey("Metric value should be correctly set", func() {
					So(mts[0].Data(), ShouldEqual, 43)
				})
			})

		})
	})
}

func TestGetMetrics(t *testing.T) {
	Convey("Given docker id and running containers info", t , func() {
		longDockerId := "1234567890ab9207edb4e6188cf5be3294c23c936ca449c3d48acd2992e357a8"
		containersInfo := []ContainerInfo{ContainerInfo{Id: longDockerId}}
		mountPoint := "cgroup/mount/point/path"
		stats := cgroups.NewStats()

		Convey("and docker plugin initialized", func() {
			mockClient := new(ClientMock)
			mockStats := new(StatsMock)
			mockTools := new(ToolsMock)
			mockWrapper := map[string]Stats{"cpu": mockStats}

			mockTools.On(
				"Map2Namespace", mock.Anything,	mock.AnythingOfType("string"), mock.AnythingOfType("*[]string")).Return().Run(
				func(args mock.Arguments) {
					id := args.String(1)
					ns := args.Get(2).(*[]string)

					*ns = append(*ns, filepath.Join(id[:12], "cpu_stats/cpu_usage/total_usage"))
				})

			mockClient.On("FindCgroupMountpoint", "cpu").Return(mountPoint, nil)
			mockStats.On("GetStats", mock.AnythingOfType("string"), mock.AnythingOfType("*cgroups.Stats")).Return(nil)

			d := &docker{
				stats: 				stats,
				client: 			mockClient,
				tools: 				mockTools,
				groupWrap: 			mockWrapper,
				containersInfo: 	containersInfo,
				hostname: 			"",
			}

			Convey("When GetMetrics is called", func() {
				mts, err := d.GetMetrics()

				Convey("Then no error should be reported", func() {
					So(err, ShouldBeNil)
				})

				Convey("Then one metric should be returned", func() {
					So(len(mts), ShouldEqual, 1)
				})

				Convey("Then metric namespace should be correctly set", func() {
					ns := filepath.Join(mts[0].Namespace()...)
					expected := filepath.Join(
						NS_VENDOR, NS_CLASS, NS_PLUGIN, longDockerId[:12], "cpu_stats", "cpu_usage", "total_usage")
					So(ns, ShouldEqual, expected)
				})
			})
		})

	})
}