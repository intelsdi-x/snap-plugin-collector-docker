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
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func parseEntry(line string) (name string, value uint64, err error) {
	fields := strings.Fields(line)
	if len(fields) != 2 {
		return name, value, fmt.Errorf("Invalid format: %s", line)
	}

	value, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return name, value, err
	}

	return fields[0], value, nil
}

func parseIntValue(file string) (uint64, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)

}

func parseStrValue(file string) (string, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(raw)), nil

}
