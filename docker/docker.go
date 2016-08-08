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

package docker

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-collector-docker/client"
	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"

	dock "github.com/fsouza/go-dockerclient"
	utils "github.com/intelsdi-x/snap-plugin-utilities/ns"
)

const (
	// namespace vendor prefix
	NS_VENDOR = "intel"
	// namespace plugin name
	NS_PLUGIN = "docker"
	// version of plugin
	VERSION = 4

	// each metric starts with prefix "/intel/docker/<docker_id>"
	lengthOfNsPrefix = 3
)

type containerData struct {
	ID string `json:"-"`
	// Basic info about the container (status, creation time, image name, etc.)
	Info wrapper.Specification `json:"spec"`

	// Container's statistics (cpu usage, memory usage, network stats, etc.)
	Stats *wrapper.Statistics `json:"stats"`
}

type docker struct {
	containers map[string]containerData      // holds data for a container under its short id
	client     client.DockerClientInterface  // client for communication with docker (basic info, stats, mount points)
	list       map[string]dock.APIContainers // contain list of all available docker containers with info about their specification

}

// dynamicElement is defined by its name and description
type dynamicElement struct {
	name        string
	description string
}

type nsCreator struct {
	dynamicElements map[string]dynamicElement
}

// definedDynamicElements holds expected dynamic element(s) with definition in docker metrics namespaces which occurs after the key-word
var definedDynamicElements = map[string]dynamicElement{
	"filesystem":   dynamicElement{"device_name", "a name of filesystem device"},
	"labels":       dynamicElement{"label_key", "a key of container's label"},
	"network":      dynamicElement{"network_interface", "a name of network interface or 'total' for aggregate"},
	"percpu_usage": dynamicElement{"cpu_id", "an id of cpu"},
}

// New returns initialized docker plugin or error if failed to connect to docker deamon
func New() (*docker, error) {
	dc, err := client.NewDockerClient()
	if err != nil {
		return nil, err
	}

	return &docker{
		containers: map[string]containerData{},
		client:     dc,
		list:       map[string]dock.APIContainers{},
	}, nil
}

// CollectMetrics retrieves values of requested metrics
func (d *docker) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	var err error
	metrics := []plugin.MetricType{}
	d.list = map[string]dock.APIContainers{}

	// get list of possible network metrics
	networkMetrics := []string{}
	utils.FromCompositionTags(wrapper.NetworkInterface{}, "", &networkMetrics)

	// get list of all running containers
	d.list, err = d.client.ListContainersAsMap()
	if err != nil {
		fmt.Fprintln(os.Stderr, "The list of running containers cannot be retrived, err=", err)
		return nil, err
	}

	// retrieve requested docker ids
	rids, err := d.getRequestedIDs(mts...)
	if err != nil {
		return nil, err
	}

	// for each requested id set adequate item into docker.container struct with stats
	for _, rid := range rids {

		if contSpec, exist := d.list[rid]; !exist {
			return nil, fmt.Errorf("Docker container does not exist, container_id=%s", rid)
		} else {
			stats, err := d.client.GetStatsFromContainer(contSpec.ID, true)
			if err != nil {
				return nil, err
			}

			// set new item to docker.container structure
			d.containers[rid] = containerData{
				ID: contSpec.ID,
				Info: wrapper.Specification{
					Status:     contSpec.Status,
					Created:    time.Unix(contSpec.Created, 0).Format("2006-01-02T15:04:05Z07:00"),
					Image:      contSpec.Image,
					SizeRw:     contSpec.SizeRw,
					SizeRootFs: contSpec.SizeRootFs,
					Labels:     contSpec.Labels,
				},
				Stats: stats,
			}

		}
	}

	for _, mt := range mts {
		ids, err := d.getRequestedIDs(mt)
		if err != nil {
			return nil, err
		}

		for _, id := range ids {
			ns := make([]core.NamespaceElement, len(mt.Namespace()))
			copy(ns, mt.Namespace())
			ns[2].Value = id

			// omit "spec" metrics for root
			if id == "root" && mt.Namespace()[lengthOfNsPrefix].Value == "spec" {
				continue
			}
			isDynamic, indexes := mt.Namespace()[lengthOfNsPrefix:].IsDynamic()

			metricName := mt.Namespace().Strings()[lengthOfNsPrefix:]

			// remove added static element (`value`)
			if metricName[len(metricName)-1] == "value" {
				metricName = metricName[:len(metricName)-1]
			}

			if !isDynamic {
				metric := plugin.MetricType{
					Timestamp_: time.Now(),
					Namespace_: ns,
					Data_:      utils.GetValueByNamespace(d.containers[id], metricName),
					Tags_:      mt.Tags(),
					Config_:    mt.Config(),
					Version_:   VERSION,
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
					for deviceName := range d.containers[id].Stats.Filesystem {
						devices = append(devices, deviceName)
					}
				} else {
					// device name is requested explicitly
					device := metricName[0]
					if _, ok := d.containers[id].Stats.Filesystem[device]; !ok {
						return nil, fmt.Errorf("In metric %s the given device name is invalid (no stats for this device)", mt.Namespace().String())
					}

					devices = append(devices, metricName[0])
				}

				for _, device := range devices {
					rns := make([]core.NamespaceElement, len(ns))
					copy(rns, ns)

					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = device

					metric := plugin.MetricType{
						Timestamp_: time.Now(),
						Namespace_: rns,
						Data_:      utils.GetValueByNamespace(d.containers[id].Stats.Filesystem[device], metricName[1:]),
						Tags_:      mt.Tags(),
						Config_:    mt.Config(),
						Version_:   VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "labels":
				// get docker labels
				labelKeys := []string{}
				if metricName[0] == "*" {
					// when label key is requested as an asterisk - take all available labels
					for labelKey := range d.containers[id].Info.Labels {
						labelKeys = append(labelKeys, labelKey)
					}
				} else {
					labelKey := metricName[0]
					if _, ok := d.containers[id].Info.Labels[labelKey]; !ok {
						return nil, fmt.Errorf("In metric %s the given label is invalid (no value for this label key)", mt.Namespace().String())
					}

					labelKeys = append(labelKeys, metricName[0])
				}

				for _, labelKey := range labelKeys {
					rns := make([]core.NamespaceElement, len(ns))
					copy(rns, ns)
					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = utils.ReplaceNotAllowedCharsInNamespacePart(labelKey)
					metric := plugin.MetricType{
						Timestamp_: time.Now(),
						Namespace_: rns,
						Data_:      d.containers[id].Info.Labels[labelKey],
						Tags_:      mt.Tags(),
						Config_:    mt.Config(),
						Version_:   VERSION,
					}

					metrics = append(metrics, metric)
				}

			case "network":
				//get docker network tx/rx statistics
				netInterfaces := []string{}
				ifaceMap := map[string]wrapper.NetworkInterface{}
				for _, iface := range d.containers[id].Stats.Network {
					ifaceMap[iface.Name] = iface
				}

				// support wildcard on interface name
				if metricName[0] == "*" {
					for _, netInterface := range d.containers[id].Stats.Network {
						netInterfaces = append(netInterfaces, netInterface.Name)
					}
				} else {
					netInterface := metricName[0]
					if _, ok := ifaceMap[netInterface]; !ok {
						return nil, fmt.Errorf("In metric %s the given network interface is invalid (no stats for this net interface)", mt.Namespace().String())
					}
					netInterfaces = append(netInterfaces, metricName[0])
				}

				for _, ifaceName := range netInterfaces {
					rns := make([]core.NamespaceElement, len(ns))
					copy(rns, ns)
					rns[indexOfDynamicElement+lengthOfNsPrefix].Value = ifaceName
					metric := plugin.MetricType{
						Timestamp_: time.Now(),
						Namespace_: rns,
						Data_:      utils.GetValueByNamespace(ifaceMap[ifaceName], metricName[1:]),
						Tags_:      mt.Tags(),
						Config_:    mt.Config(),
						Version_:   VERSION,
					}
					metrics = append(metrics, metric)
				}

			case "percpu_usage":
				numOfCPUs := len(d.containers[id].Stats.CgroupStats.CpuStats.CpuUsage.PercpuUsage) - 1
				if metricName[0] == "*" {
					// when cpu ID is requested as an asterisk - take all available
					for cpuID, val := range d.containers[id].Stats.CgroupStats.CpuStats.CpuUsage.PercpuUsage {
						rns := make([]core.NamespaceElement, len(ns))
						copy(rns, ns)

						rns[indexOfDynamicElement+lengthOfNsPrefix].Value = strconv.Itoa(cpuID)

						metric := plugin.MetricType{
							Timestamp_: time.Now(),
							Namespace_: rns,
							Data_:      val,
							Tags_:      mt.Tags(),
							Config_:    mt.Config(),
							Version_:   VERSION,
						}
						metrics = append(metrics, metric)
					}
				} else {
					cpuID, err := strconv.Atoi(metricName[0])
					if err != nil {
						return nil, fmt.Errorf("In metric %s the given cpu id is invalid, err=%v", mt.Namespace().String(), err)
					}
					if cpuID > numOfCPUs || cpuID < 0 {
						return nil, fmt.Errorf("In metric %s the given cpu id is invalid, expected value in range 0-%d", mt.Namespace().String(), numOfCPUs)
					}

					metric := plugin.MetricType{
						Timestamp_: time.Now(),
						Namespace_: ns,
						Data_:      d.containers[id].Stats.CgroupStats.CpuStats.CpuUsage.PercpuUsage[cpuID],
						Tags_:      mt.Tags(),
						Config_:    mt.Config(),
						Version_:   VERSION,
					}
					metrics = append(metrics, metric)
				}

			} // the end of switch statsType
		} // the end of range over ids
	}

	if len(metrics) == 0 {
		return nil, errors.New("No metric found")
	}

	return metrics, nil
}

// GetConfigPolicy returns plugin config policy
func (d *docker) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}

// GetMetricTypes returns list of available metrics
func (d *docker) GetMetricTypes(_ plugin.ConfigType) ([]plugin.MetricType, error) {
	var metricTypes []plugin.MetricType
	var err error

	// initialize containerData struct
	data := containerData{
		Stats: wrapper.NewStatistics(),
	}

	// generate available namespace for data container structure
	dockerMetrics := []string{}
	utils.FromCompositeObject(data, "", &dockerMetrics)
	nscreator := nsCreator{dynamicElements: definedDynamicElements}

	for _, metricName := range dockerMetrics {

		ns := core.NewNamespace(NS_VENDOR, NS_PLUGIN).
			AddDynamicElement("docker_id", "an id of docker container")

		if ns, err = nscreator.createMetricNamespace(ns, metricName); err != nil {
			// skip this metric name which is not supported
			// fmt.Fprintf(os.Stderr, "Error in creating metric namespace: %v\n", err)
			continue
		}
		metricType := plugin.MetricType{
			Namespace_: ns,
			Version_:   VERSION,
		}
		metricTypes = append(metricTypes, metricType)
	}

	return metricTypes, nil
}

// createMetricNamespace returns metric namespace based on given `ns` which is used as a prefix; all dynamic elements
// in the `metricName` are defined based on content of map `dynamicElements`
func (creator *nsCreator) createMetricNamespace(ns core.Namespace, metricName string) (core.Namespace, error) {
	metricName = strings.TrimSpace(metricName)

	if len(metricName) == 0 {
		return nil, errors.New("Cannot create metric namespace: empty metric name")
	}

	elements := strings.Split(metricName, "/")

	// check if metricName contains only static elements
	if !strings.Contains(metricName, "*") {
		ns = ns.AddStaticElements(elements...)
		return ns, nil
	}

	// when metric name contains dynamic element iterate over elements
	for index, element := range elements {
		if element == "*" {
			// the following element is dynamic
			dynamicElement, ok := creator.dynamicElements[elements[index-1]]
			// check if this dynamic element is supported (name and description are available)
			if !ok {
				return nil, fmt.Errorf("Unknown dynamic element in metric `%s` under index %d", metricName, index)
			}
			// add recognize dynamic element (define its name and description)
			ns = ns.AddDynamicElement(dynamicElement.name, dynamicElement.description)

			if len(elements)-1 == index {
				// in case when an asterisk is the last element, add `value` at the end of ns
				ns = ns.AddStaticElement("value")
			}
		} else {
			// the following element is static
			ns = ns.AddStaticElement(element)
		}
	}
	if len(ns) == 0 {
		return nil, fmt.Errorf("Cannot create metric namespace for metric %s", metricName)
	}
	return ns, nil
}

// availableContainer returns IDs of all available docker containers
func (d *docker) availableContainers() []string {
	ids := []string{}

	// iterate over list of available dockers
	for id := range d.list {
		ids = append(ids, id)
	}

	return ids
}

func appendIfMissing(items []string, newItem string) []string {
	for _, item := range items {
		if newItem == item {
			// do not append new item
			return items
		}
	}
	return append(items, newItem)
}

// getRequestedIDs returns requested docker ids
func (d *docker) getRequestedIDs(mt ...plugin.MetricType) ([]string, error) {
	rids := []string{}
	for _, m := range mt {
		ns := m.Namespace().Strings()
		if ok := validateMetricNamespace(ns); !ok {
			return nil, fmt.Errorf("Invalid name of metric %+s", m.Namespace().String())
		}

		rid := m.Namespace().Strings()[2]
		if rid == "*" {
			// all available dockers are requested
			rids := d.availableContainers()
			rids = appendIfMissing(rids, "root")
			return rids, nil
		} else if rid == "root" {
			rids = appendIfMissing(rids, "root")
			continue
		}

		shortID, err := client.GetShortID(rid)
		if err != nil {
			return nil, err
		}

		if !d.validateDockerID(shortID) {
			return nil, fmt.Errorf("Docker container %+s cannot be found", rid)
		}

		rids = appendIfMissing(rids, shortID)
	}

	if len(rids) == 0 {
		return nil, errors.New("Cannot retrieve requested docker id")
	}
	return rids, nil
}

// validateDockerID returns true if docker with a given dockerID has been found on list of available dockers
func (d *docker) validateDockerID(dockerID string) bool {
	if _, exist := d.list[dockerID]; exist {
		return true
	}
	return false
}

// validateMetricNamespace returns true if the given metric namespace has the required length
func validateMetricNamespace(ns []string) bool {
	// metric namespace has to contain the following elements:
	//  the prefix, metric_type (spec,cgroups or network) and metric_name
	if len(ns) < lengthOfNsPrefix+2 {
		return false
	}
	return true
}
