//
// +build small

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
	"errors"
	"fmt"
	"strings"
	"testing"

	dock "github.com/fsouza/go-dockerclient"
	"github.com/intelsdi-x/snap-plugin-collector-docker/client"
	. "github.com/intelsdi-x/snap-plugin-collector-docker/mocks"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

var mockStats = CreateMockStats()
var mockDockerID = "a26c852ce22c"
var mockDockerHost = "root"

var mockListOfContainers = map[string]dock.APIContainers{
	mockDockerHost: dock.APIContainers{
		ID: "/",
	},
	mockDockerID: dock.APIContainers{
		ID:         "a26c852ce22cbf94f75299b879ccb0d94427aa265778e1e9d6e6483ffb7837ed",
		Image:      "my/image:latest",
		Command:    "my-command.sh",
		Created:    1469187756,
		Status:     "Up 4 weeks",
		SizeRw:     0,
		SizeRootFs: 0,
		Names:      []string{"/naught_goodall"},
		Labels: map[string]string{
			"lkey1": "lval1",
			"lkey2": "lval2",
			"lkey3": "lval3",
		},
	},
}

var mockMts = []plugin.Metric{
	// representation of metrics grouped as `spec`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("spec", "creation_time"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	// representation of metrics grouped as `cgroup/cpu_stats`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "cgroups", "cpu_stats", "cpu_usage", "total_usage"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "cgroups", "cpu_stats", "cpu_usage", "percpu_usage").
			AddDynamicElement("cpu_id", "an id of cpu").
			AddStaticElement("value"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},

	// representation of metrics grouped as `cgroups/memory_stats`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "cgroups", "memory_stats", "cache"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "cgroups", "memory_stats", "stats", "pgpgin"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "cgroups", "memory_stats", "usage", "max_usage"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},

	// representation of metrics grouped as `connection`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "connection", "tcp", "established"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "connection", "tcp6", "established"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},

	// representation of metrics grouped as `filesystem`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "filesystem").
			AddDynamicElement("device_name", "a name of filesystem device").
			AddStaticElement("usage"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},

	// representation of metrics grouped as `network`
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "network").
			AddDynamicElement("network_interface", "a name of network interface or 'total' for aggregate").
			AddStaticElement("rx_bytes"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("stats", "network").
			AddDynamicElement("network_interface", "a name of network interface or 'total' for aggregate").
			AddStaticElement("tx_bytes"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container").
			AddStaticElements("spec", "labels").
			AddDynamicElement("label_key", "a key of container's label").
			AddStaticElement("value"),
		Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
	},
}

func TestGetMetricTypes(t *testing.T) {
	Convey("create Docker collector plugin", t, func() {

		mc := new(ClientMock)
		mc.On("NewDockerClient").Return(&client.DockerClient{}, nil)
		dockerPlg := &docker{
			containers: map[string]containerData{},
			client:     mc,
		}

		Convey("get list of available metrics", func() {
			metrics, err := dockerPlg.GetMetricTypes(plugin.Config{})
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeEmpty)

			Convey("check if version is set ", func() {
				for _, metric := range metrics {
					So(metric.Version, ShouldEqual, VERSION)
				}
			})
		})
	})
}

func TestGetConfigPolicy(t *testing.T) {
	Convey("create Docker collector plugin", t, func() {
		mc := new(ClientMock)
		mc.On("NewDockerClient").Return(&client.DockerClient{}, nil)
		dockerPlg := &docker{
			containers: map[string]containerData{},
			client:     mc,
		}

		Convey("get config policy", func() {
			configPolicy, err := dockerPlg.GetConfigPolicy()
			So(err, ShouldBeNil)
			So(configPolicy, ShouldNotBeNil)
		})
	})
}

func TestCollectMetrics(t *testing.T) {
	dockerPlg := &docker{
		containers: map[string]containerData{},
	}

	Convey("return an error when there is no available container", t, func() {
		mc := new(ClientMock)
		mc.On("ListContainersAsMap").Return(nil, errors.New("No docker container found"))
		dockerPlg.client = mc
		metrics, err := dockerPlg.CollectMetrics(mockMts)
		So(err, ShouldNotBeNil)
		So(metrics, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "No docker container found")
	})
	Convey("return an error when cannot get statistics", t, func() {
		mc := new(ClientMock)
		mc.On("ListContainersAsMap").Return(mockListOfContainers, nil)
		mc.On("GetStatsFromContainer").Return(nil, errors.New("Cannot get statistics"))
		dockerPlg.client = mc
		metrics, err := dockerPlg.CollectMetrics(mockMts)
		So(err, ShouldNotBeNil)
		So(metrics, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "Cannot get statistics")
	})
	Convey("successful collect metrics", t, func() {
		mc := new(ClientMock)
		mc.On("ListContainersAsMap").Return(mockListOfContainers, nil)
		mc.On("GetStatsFromContainer").Return(mockStats, nil)

		dockerPlg.client = mc
		metrics, err := dockerPlg.CollectMetrics(mockMts)
		So(err, ShouldBeNil)
		So(metrics, ShouldNotBeEmpty)

		// each of collected metrics should have set a version of collector plugin
		Convey("collected metrics have set plugin version", func() {
			for _, metric := range metrics {
				So(metric.Version, ShouldEqual, VERSION)
			}
		})

		// collected metrics should not contain an asterisk in a namespace
		Convey("collected metrics have specified namespace", func() {
			for _, metric := range metrics {
				So(metric.Namespace.Strings(), ShouldNotContain, "*")
			}
		})

	})
	Convey("successful collect metrics for specified dynamic metric", t, func() {
		mc := new(ClientMock)
		mc.On("ListContainersAsMap").Return(mockListOfContainers, nil)
		mc.On("GetStatsFromContainer").Return(mockStats, nil)

		dockerPlg.client = mc

		Convey("for specific dynamic elements: docker_id", func() {
			mockMt := plugin.Metric{
				Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
					AddDynamicElement("docker_id", "an id of docker container").
					AddStaticElements("stats", "cgroups", "memory_stats", "cache"),
				Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
			}

			Convey("succefull when specified container exists", func() {
				Convey("for short docker_id", func() {
					// specify docker id of requested metric type as a short
					mockMt.Namespace[2].Value = mockDockerID

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldBeNil)
					So(metrics, ShouldNotBeEmpty)
					So(len(metrics), ShouldEqual, 1)
					So(metrics[0].Namespace, ShouldResemble, mockMt.Namespace)
				})
				Convey("for long docker_id", func() {
					// specify docker id of requested metric type as a long
					mockMt.Namespace[2].Value = mockListOfContainers[mockDockerID].ID

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldBeNil)
					So(metrics, ShouldNotBeEmpty)
					So(len(metrics), ShouldEqual, 1)
					So(strings.Join(metrics[0].Namespace.Strings(), "/"), ShouldEqual, "intel/docker/"+mockDockerID+"/stats/cgroups/memory_stats/cache")
				})
				Convey("for host of docker_id", func() {
					// specify docker id of requested metric type
					mockMt.Namespace[2].Value = "root"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldBeNil)
					So(metrics, ShouldNotBeEmpty)
					So(len(metrics), ShouldEqual, 1)
					So(strings.Join(metrics[0].Namespace.Strings(), "/"), ShouldEqual, "intel/docker/root/stats/cgroups/memory_stats/cache")
				})
			})
			Convey("return an error when specified docker_id is invalid", func() {
				Convey("when there is no such container", func() {
					// specify id (12 chars) of docker container which not exist (it's not returned by ListContainerAsMap())
					mockMt.Namespace[2].Value = "111111111111"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldNotBeNil)
					So(metrics, ShouldBeEmpty)
					So(err.Error(), ShouldEqual, "Docker container 111111111111 cannot be found")
				})
				Convey("when specified docker_id has invalid format", func() {
					mockMt := plugin.Metric{
						Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
							AddDynamicElement("docker_id", "an id of docker container").
							AddStaticElements("stats", "cgroups", "memory_stats", "cache"),
						Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
					}
					// specify requested docker id in invalid way (shorter than 12 chars)
					mockMt.Namespace[2].Value = "1"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldNotBeNil)
					So(metrics, ShouldBeEmpty)
					So(err.Error(), ShouldEqual, "Docker id 1 is too short (the length of id should equal at least 12)")
				})
			})
		})
		Convey("for specific dynamic elements: docker_id and cpu_id", func() {
			mockMt := plugin.Metric{
				Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
					AddDynamicElement("docker_id", "an id of docker container").
					AddStaticElements("stats", "cgroups", "cpu_stats", "cpu_usage", "percpu_usage").
					AddDynamicElement("cpu_id", "an id of cpu").
					AddStaticElement("value"),
				Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
			}
			// specify docker_id and cpu_id of requested metric type
			mockMt.Namespace[2].Value = mockDockerID

			Convey("successful when specified cpu_id is valid", func() {
				mockMt.Namespace[8].Value = "0"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldBeNil)
				So(metrics, ShouldNotBeEmpty)
				So(len(metrics), ShouldEqual, 1)
				So(metrics[0].Namespace, ShouldResemble, mockMt.Namespace)
			})
			Convey("return an error when specified cpu_id is invalid", func() {
				Convey("when cpu_id is out of range", func() {
					// specify cpu_id which does not exist (out of range)
					mockMt.Namespace[8].Value = "100"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldNotBeNil)
					So(metrics, ShouldBeEmpty)
				})
				Convey("when cpu_id is negative", func() {
					mockMt.Namespace[8].Value = "-1"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldNotBeNil)
					So(metrics, ShouldBeEmpty)
				})
				Convey("when cpu_id is a float", func() {
					mockMt.Namespace[8].Value = "1.0"

					metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
					So(err, ShouldNotBeNil)
					So(metrics, ShouldBeEmpty)
				})
			})
		})
		Convey("for specific dynamic elements: docker_id and device_name", func() {
			mockMt := plugin.Metric{
				Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
					AddDynamicElement("docker_id", "an id of docker container").
					AddStaticElements("stats", "filesystem").
					AddDynamicElement("device_name", "a name of filesystem device").
					AddStaticElement("usage"),
				Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
			}
			mockMt.Namespace[2].Value = mockDockerID

			Convey("successful when specified device exists", func() {
				mockMt.Namespace[5].Value = "dev1"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldBeNil)
				So(metrics, ShouldNotBeEmpty)
				So(len(metrics), ShouldEqual, 1)
				So(metrics[0].Namespace, ShouldResemble, mockMt.Namespace)
			})
			Convey("return an error when specified device is invalid", func() {
				mockMt.Namespace[5].Value = "invalid_dev_name"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldNotBeNil)
				So(metrics, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, fmt.Sprintf("In metric %s the given device name is invalid (no stats for this device)", strings.Join(mockMt.Namespace.Strings(), "/")))
			})
		})
		Convey("for specific dynamic elements: docker_id and network_interface", func() {
			mockMt := plugin.Metric{
				Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
					AddDynamicElement("docker_id", "an id of docker container").
					AddStaticElements("stats", "network").
					AddDynamicElement("network_interface", "a name of network interface or 'total' for aggregate").
					AddStaticElement("rx_bytes"),
				Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
			}
			// specify docker_id and device_name of requested metric type
			mockMt.Namespace[2].Value = mockDockerID

			Convey("successful when specified network interface exists", func() {
				mockMt.Namespace[5].Value = "eth0"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldBeNil)
				So(metrics, ShouldNotBeEmpty)
				So(len(metrics), ShouldEqual, 1)
				So(metrics[0].Namespace, ShouldResemble, mockMt.Namespace)
			})
			Convey("return an error when specified network interface is invalid", func() {
				mockMt.Namespace[5].Value = "eth0_invalid"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldNotBeNil)
				So(metrics, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, fmt.Sprintf("In metric %s the given network interface is invalid (no stats for this net interface)", strings.Join(mockMt.Namespace.Strings(), "/")))
			})
		})
		Convey("for specific dynamic elements: docker_id and label_key", func() {
			mockMt := plugin.Metric{
				Namespace: plugin.NewNamespace(NS_VENDOR, NS_PLUGIN).
					AddDynamicElement("docker_id", "an id of docker container").
					AddStaticElements("spec", "labels").
					AddDynamicElement("label_key", "a key of container's label").
					AddStaticElement("value"),
				Config: plugin.Config{"endpoint": "unix:///var/run/docker.sock"},
			}
			// specify docker_id and device_name of requested metric type
			mockMt.Namespace[2].Value = mockDockerID

			Convey("successful when specified label exists", func() {
				mockMt.Namespace[5].Value = "lkey1"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldBeNil)
				So(metrics, ShouldNotBeEmpty)
				So(len(metrics), ShouldEqual, 1)
				So(metrics[0].Namespace, ShouldResemble, mockMt.Namespace)
			})
			Convey("return an error when specified label is invalid (not exist)", func() {
				mockMt.Namespace[5].Value = "lkey1_invalid"

				metrics, err := dockerPlg.CollectMetrics([]plugin.Metric{mockMt})
				So(err, ShouldNotBeNil)
				So(metrics, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, fmt.Sprintf("In metric %s the given label is invalid (no value for this label key)", strings.Join(mockMt.Namespace.Strings(), "/")))
			})
		})
	})
}

func TestCreateMetricNamespace(t *testing.T) {
	Convey("create metric namespace", t, func() {
		nscreator := nsCreator{}

		Convey("return an error when metric name is empty", func() {
			ns, err := nscreator.createMetricNamespace(plugin.NewNamespace(), "")
			So(err, ShouldNotBeNil)
			So(ns, ShouldBeNil)
		})

		Convey("when metric name contains only static elements", func() {
			ns, err := nscreator.createMetricNamespace(plugin.NewNamespace("vendor", "plugin"), "disk/total_usage")
			So(err, ShouldBeNil)
			So(ns, ShouldNotBeNil)
			So(strings.Join(ns.Strings(), "/"), ShouldEqual, "vendor/plugin/disk/total_usage")
		})

		Convey("when metric name contains dynamic element", func() {

			Convey("return an error for unknown dynamic element", func() {
				ns, err := nscreator.createMetricNamespace(plugin.NewNamespace("vendor", "plugin"), "disk/*/usage")
				So(err, ShouldNotBeNil)
				So(ns, ShouldBeNil)
				So(err.Error(), ShouldEqual, "Unknown dynamic element in metric `disk/*/usage` under index 1")
			})

			Convey("successful create metric namespace with dynamic element", func() {
				// set definition of dynamic element (its name and description)
				nscreator.dynamicElements = map[string]dynamicElement{
					"disk": dynamicElement{"disk_id", "id of disk"},
				}
				ns, err := nscreator.createMetricNamespace(plugin.NewNamespace("vendor", "plugin"), "disk/*/usage")
				So(err, ShouldBeNil)
				So(ns, ShouldNotBeEmpty)
				So(strings.Join(ns.Strings(), "/"), ShouldEqual, "vendor/plugin/disk/*/usage")
				So(ns.Element(3).Description, ShouldEqual, nscreator.dynamicElements["disk"].description)
				So(ns.Element(3).Name, ShouldEqual, nscreator.dynamicElements["disk"].name)
			})

			Convey("successful create metric namespace with dynamic element at the end of metric name", func() {
				nscreator.dynamicElements = map[string]dynamicElement{
					"percpu_usage": dynamicElement{"cpu_id", "id of cpu"},
				}
				ns, err := nscreator.createMetricNamespace(plugin.NewNamespace("vendor", "plugin"), "percpu_usage/*")
				So(err, ShouldBeNil)
				So(ns, ShouldNotBeEmpty)
				// metric namespace should not end with an asterisk,
				// so element `value` is expected to be added
				So(strings.Join(ns.Strings(), "/"), ShouldEqual, "vendor/plugin/percpu_usage/*/value")
				So(ns.Element(3).Description, ShouldEqual, nscreator.dynamicElements["percpu_usage"].description)
				So(ns.Element(3).Name, ShouldEqual, nscreator.dynamicElements["percpu_usage"].name)
			})

		})

	})

}
