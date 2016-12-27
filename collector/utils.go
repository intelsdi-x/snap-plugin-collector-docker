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
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

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
	"filesystem":                 {"device_name", "a name of filesystem device"},
	"labels":                     {"label_key", "a key of container's label"},
	"network":                    {"network_interface", "a name of network interface or 'total' for aggregate"},
	"per_cpu":                    {"cpu_id", "an id of cpu"},
	"io_service_bytes_recursive": {"device_name", "a name of block device"},
	"io_serviced_recursive":      {"device_name", "a name of block device"},
	"io_queue_recursive":         {"device_name", "a name of block device"},
	"io_service_time_recursive":  {"device_name", "a name of block device"},
	"io_wait_time_recursive":     {"device_name", "a name of block device"},
	"io_merged_recursive":        {"device_name", "a name of block device"},
	"io_time_recursive":          {"device_name", "a name of block device"},
	"sectors_recursive":          {"device_name", "a name of block device"},
	"hugetlb_stats":              {"size", "hugetlb page size"},
}

func initClient(c *collector, endpoint string) error {
	dc, err := container.NewDockerClient(endpoint)
	if err != nil {
		return err
	}

	params, err := dc.GetDockerParams("DockerRootDir", "Driver")
	if err != nil {
		return err
	}

	c.rootDir = params["DockerRootDir"]
	c.driver = params["Driver"]
	c.client = dc

	log.WithFields(log.Fields{
		"block": "initClient",
	}).Infof("Docker client initialized with storage driver %s and docker root dir %s", c.driver, c.rootDir)

	return nil
}

// createMetricNamespace returns metric namespace based on given `ns` which is used as a prefix; all dynamic elements
// in the `metricName` are defined based on content of map `dynamicElements`
func (creator *nsCreator) createMetricNamespace(ns plugin.Namespace, metricName string) (plugin.Namespace, error) {
	metricName = strings.TrimSpace(metricName)

	if len(metricName) == 0 {
		return nil, fmt.Errorf("Cannot create metric namespace: empty metric name %s", metricName)
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

func getQueryGroup(ns []string) (string, error) {
	if ns[0] == "spec" {
		return ns[0], nil
	}

	for _, ne := range ns {
		if _, exists := getters[ne]; exists {
			return ne, nil
		}
	}
	return "", fmt.Errorf("Cannot identify query group for given namespace %s", strings.Join(ns, "/"))
}

func appendIfMissing(collectGroup map[string]map[string]struct{}, rid string, query string) {
	group, exists := collectGroup[rid]
	if !exists {
		collectGroup[rid] = map[string]struct{}{query: {}}
	}

	if _, exists := group[query]; !exists {
		collectGroup[rid][query] = struct{}{}
	}
}

func getDockerConfig(cfg plugin.Config) (map[string]string, error) {
	config := make(map[string]string)
	values := []string{"endpoint", "procfs"}
	var err error
	for _, v := range values {
		config[v], err = cfg.GetString(v)
		if err != nil {
			return config, err
		}
	}
	return config, nil
}
