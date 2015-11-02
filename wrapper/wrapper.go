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

package wrapper

import (
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/fs"
)

var Cgroups2Stats = map[string]Stats{
	"cpuset": &fs.CpusetGroup{},
	"cpu": &fs.CpuGroup{},
	"cpuacct": &fs.CpuacctGroup{},
	"memory": &fs.MemoryGroup{},
	"devices": &fs.DevicesGroup{},
	"freezer": &fs.FreezerGroup{},
	"net_cls": &fs.NetClsGroup{},
	"blkio": &fs.BlkioGroup{},
	"perf_event": &fs.PerfEventGroup{},
	"net_prio": &fs.NetPrioGroup{},
	"hugetlb": &fs.HugetlbGroup{},
}

type Stats interface {
	GetStats(path string, stats *cgroups.Stats) error
}