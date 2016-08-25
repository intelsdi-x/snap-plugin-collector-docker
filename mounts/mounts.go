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

// Package mounts provides information about mountpoints
package mounts

import (
	"os"
	"strings"
)

const (
	procfsMountEnv     = "PROCFS_MOUNT"
	procfsMountDefault = "/proc"
)

// ProcfsMountPoint holds a path to mountpoint of procfs, defaults to `/proc`
var ProcfsMountPoint = getProcfsMountpoint()

func getProcfsMountpoint() string {
	if procfsMount := os.Getenv(procfsMountEnv); procfsMount != "" {
		//trim suffix in case that env var contains slash in the end
		return strings.TrimSuffix(procfsMount, "/")
	}
	return procfsMountDefault
}
