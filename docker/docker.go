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
	"os"
	"fmt"
	"time"
	"strings"
	"path/filepath"
	"encoding/json"

	"github.com/vektra/errors"
	"github.com/opencontainers/runc/libcontainer/cgroups"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse-plugin-collector-docker/client"
	"github.com/intelsdi-x/pulse-plugin-collector-docker/wrapper"
	tls "github.com/intelsdi-x/pulse-plugin-collector-docker/tools"
)

const (
	// namespace vendor prefix
	NS_VENDOR = "intel"
	// namespace os prefix
	NS_CLASS = "linux"
	// namespace plugin name
	NS_PLUGIN = "docker"
	// version of plugin
	VERSION = 1
	// mount info
	mountInfo = "/proc/self/mountinfo"
)

type docker struct {
	stats 			*cgroups.Stats					// structure for stats storage
	client			client.DockerClientInterface	// client for communication with docker (basic info, mount points)
	tools			tls.ToolsInterface				// tools for handling namespaces and processing stats
	containersInfo 	[]client.ContainerInfo			// basic info about running containers
	groupWrap		map[string]wrapper.Stats		// wrapper for cgroup name and interface for stats extraction
	hostname 		string							// name of the host
}

func NewDocker() (*docker, error) {
	host, _ := os.Hostname()
	// create new docker client
	dockerClient := client.NewDockerClient()
	// list all running containers
	containers, err := dockerClient.ListContainers()

	if err != nil {
		return nil, err
	}

	d := &docker{
		stats:      		cgroups.NewStats(),
		client:         	dockerClient,
		tools:				new(tls.MyTools),
		containersInfo: 	containers,
		groupWrap:			wrapper.Cgroups2Stats,
		hostname:           host}

	return d, nil
}

func (d *docker) getStats(id string) error {

	for cg, stat := range d.groupWrap {
		// find mount point for each cgroup
		mp, err := d.client.FindCgroupMountpoint(cg)

		if err != nil {
			fmt.Printf("[WARNING] Could not find mount point for %s\n", cg)
			continue
		}

		// create path to cgroup for given docker id
		groupPath := filepath.Join(mp, "docker", id)
		// get cgroup stats for given docker
		if err := stat.GetStats(groupPath, d.stats); err != nil {
			return err
		}
	}

	return nil
}

func (d *docker) extendDockerId(shortId string) (string, error) {

	for _, cinfo := range d.containersInfo {
		if strings.HasPrefix(cinfo.Id, shortId) {
			return cinfo.Id, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Could not find long docker id for %s\n", shortId))
}

func (d *docker) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {

	metrics := make([]plugin.PluginMetricType, len(metricTypes))

	for i, metricType := range metricTypes {
		// example ns: /intel/linux/docker/31068893a2bc/cpu_stats/throttling_data/periods
		ns := metricType.Namespace()

		// extracted docker id from namespace is extended to long one
		dockerId, err := d.extendDockerId(ns[3])
		if err != nil {
			return nil, err
		}

		// long id is required to get stats for docker
		if err := d.getStats(dockerId); err != nil {
			return nil, err
		}

		// only "cgroup.Stats" part of namespace is sent to retrieve value (cpu_stats/throttling_data/periods)
		metrics[i].Data_ = d.tools.GetValueByNamespace(d.stats, ns[4:])
		metrics[i].Timestamp_ = time.Now()
		metrics[i].Namespace_ = ns
		metrics[i].Version_ = VERSION
		metrics[i].Source_ = filepath.Join(d.hostname, ns[3])
	}

	return metrics, nil
}

func (d *docker) GetMetrics() ([]plugin.PluginMetricType, error) {

	var namespaces []string
	var metricTypes []plugin.PluginMetricType

	for _, container := range d.containersInfo {
		id, err := d.extendDockerId(container.Id)
		if err != nil {
			return nil, err
		}
		// calling getStats will populate stats object
		// parsing it one will get info on available namespace
		_ = d.getStats(id)

		// marshal-unmarshal to get map with json tags as keys
		jsondata, _ := json.Marshal(d.stats)
		var jmap map[string]interface{}
		_ = json.Unmarshal(jsondata, &jmap)

		// parse map to get namespace strings
		d.tools.Map2Namespace(jmap, container.Id, &namespaces)
	}

	for _, namespace := range namespaces {
		// construct full namespace
		fullNs := filepath.Join(NS_VENDOR, NS_CLASS, NS_PLUGIN, namespace)
		metricTypes = append(metricTypes, plugin.PluginMetricType{Namespace_: strings.Split(fullNs, "/")})
	}

	return metricTypes, nil
}

func (d *docker) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return cpolicy.New(), nil
}
