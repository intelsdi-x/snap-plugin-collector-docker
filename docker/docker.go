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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/vektra/errors"

	"github.com/intelsdi-x/snap-plugin-collector-docker/client"
	tls "github.com/intelsdi-x/snap-plugin-collector-docker/tools"
	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// namespace vendor prefix
	NS_VENDOR = "intel"
	// namespace os prefix
	NS_CLASS = "linux"
	// namespace plugin name
	NS_PLUGIN = "docker"
	// version of plugin
	VERSION = 2
	// mount info
	mountInfo = "/proc/self/mountinfo"
)

// Docker plugin type
type docker struct {
	stats          *cgroups.Stats               // structure for stats storage
	client         client.DockerClientInterface // client for communication with docker (basic info, mount points)
	tools          tls.ToolsInterface           // tools for handling namespaces and processing stats
	containersInfo []client.ContainerInfo       // basic info about running containers
	groupWrap      map[string]wrapper.Stats     // wrapper for cgroup name and interface for stats extraction
	hostname       string                       // name of the host
}

// Docker plugin initializer
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
		stats:          cgroups.NewStats(),
		client:         dockerClient,
		tools:          new(tls.MyTools),
		containersInfo: containers,
		groupWrap:      wrapper.Cgroups2Stats,
		hostname:       host}

	return d, nil
}

// wrapper for cgroup stats extraction
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

// short docker id (12 char) is extended
func (d *docker) extendDockerId(shortId string) (string, error) {

	for _, cinfo := range d.containersInfo {
		if strings.HasPrefix(cinfo.Id, shortId) {
			return cinfo.Id, nil
		}
	}

	return "", errors.New(fmt.Sprintf("Could not find long docker id for %s\n", shortId))
}

func (d *docker) CollectMetrics(metricTypes []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {

	metrics := []plugin.PluginMetricType{}

	for _, metricType := range metricTypes {

		ns := metricType.Namespace()

		// wildcard for container ID
		if ns[3] == "*" {
			// example ns: /intel/linux/docker/*/cpu_stats/throttling_data/priods
			for _, container := range d.containersInfo {

				dockerId := container.Id
				// get short version of container ID
				ns[3] = dockerId[:12]

				// copy namespace so it doesn't get altered each time
				nscopy := make([]string, len(ns))
				copy(nscopy, ns)

				if err := d.getStats(dockerId); err != nil {
					return nil, err
				}

				// only "cgroup.Stats" part of namespace is sent to retrieve value (cpu_stats/throttling_data/periods)
				mt := plugin.PluginMetricType{
					Data_:      d.tools.GetValueByNamespace(d.stats, ns[4:]),
					Timestamp_: time.Now(),
					Namespace_: nscopy,
					Version_:   VERSION,
					Source_:    filepath.Join(d.hostname, ns[3]),
				}
				metrics = append(metrics, mt)
			}
		} else {
			// example ns: /intel/linux/docker/31068893a2bc/cpu_stats/throttling_data/priods

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
			mt := plugin.PluginMetricType{
				Data_:      d.tools.GetValueByNamespace(d.stats, ns[4:]),
				Timestamp_: time.Now(),
				Namespace_: ns,
				Version_:   VERSION,
				Source_:    filepath.Join(d.hostname, ns[3]),
			}
			metrics = append(metrics, mt)
		}
	}

	return metrics, nil
}

func (d *docker) GetMetricTypes(_ plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	var namespaces []string
	var metricTypes []plugin.PluginMetricType

	for _, container := range d.containersInfo {
		// calling getStats will populate stats object
		// parsing it one will get info on available namespace
		_ = d.getStats(container.Id)

		// marshal-unmarshal to get map with json tags as keys
		jsondata, _ := json.Marshal(d.stats)
		var jmap map[string]interface{}
		_ = json.Unmarshal(jsondata, &jmap)

		// parse map to get namespace strings
		d.tools.Map2Namespace(jmap, container.Id[:12], &namespaces)
	}

	// wildcard for container ID
	if len(d.containersInfo) > 0 {
		jsondata, _ := json.Marshal(d.stats)
		var jmap map[string]interface{}
		_ = json.Unmarshal(jsondata, &jmap)
		d.tools.Map2Namespace(jmap, "*", &namespaces)
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
