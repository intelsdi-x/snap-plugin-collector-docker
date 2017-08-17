// +build small

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

This file incorporates work covered by the following copyright and permission notice:
	Copyright 2014 Google Inc. All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fs

import (
	"github.com/moby/moby/pkg/mount"
)

var mockedMounts = []*mount.Info{
	{
		ID:         19,
		Parent:     25,
		Major:      0,
		Minor:      18,
		Root:       "/",
		Mountpoint: "/sys",
		Opts:       "rw,nosuid,nodev,noexec,relatime",
		Optional:   "shared:7",
		Fstype:     "sysfs",
		Source:     "sysfs",
		VfsOpts:    "rw",
	},
	{
		ID:         20,
		Parent:     25,
		Major:      0,
		Minor:      4,
		Root:       "/",
		Mountpoint: "/proc",
		Opts:       "rw,nosuid,nodev,noexec,realtime",
		Optional:   "shared:12",
		Fstype:     "proc",
		Source:     "proc",
		VfsOpts:    "rw",
	},
	{
		ID:         21,
		Parent:     25,
		Major:      0,
		Minor:      6,
		Root:       "/",
		Mountpoint: "/dev",
		Opts:       "rw,nosuid,relatime",
		Optional:   "shared:2",
		Fstype:     "devtmpfs",
		Source:     "udev",
		VfsOpts:    "rw,size=16392172k,nr_inodes=4098043,mode=755",
	},
	{
		ID:         22,
		Parent:     21,
		Major:      0,
		Minor:      14,
		Root:       "/",
		Mountpoint: "/dev/pts",
		Opts:       "rw,nosuid,noexec,relatime",
		Optional:   "shared:3",
		Fstype:     "devpts",
		Source:     "devpts",
		VfsOpts:    "rw,gid=5,mode=620,ptmxmode=000",
	},
	{
		ID:         23,
		Parent:     25,
		Major:      0,
		Minor:      19,
		Root:       "/",
		Mountpoint: "/run",
		Opts:       "rw,nosuid,noexec,relatime",
		Optional:   "shared:5",
		Fstype:     "tmpfs",
		Source:     "tmpfs",
		VfsOpts:    "rw,size=3282484k,mode=755",
	},
	{
		ID:         25,
		Parent:     0,
		Major:      8,
		Minor:      1,
		Root:       "/",
		Mountpoint: "/",
		Opts:       "rw,relatime",
		Optional:   "shared:1",
		Fstype:     "ext4",
		Source:     "/dev/sda1",
		VfsOpts:    "rw,errors=remount-ro,data=ordered",
	},
	{
		ID:         296,
		Parent:     23,
		Major:      0,
		Minor:      69,
		Root:       "/",
		Mountpoint: "/run/cgmanager/fs",
		Opts:       "rw,relatime",
		Optional:   "shared:155",
		Fstype:     "tmpfs",
		Source:     "cgmfs",
		VfsOpts:    "rw,size=100k,mode=755",
	},
	{
		ID:         142,
		Parent:     25,
		Major:      8,
		Minor:      1,
		Root:       "/tmp/var/lib/docker/aufs",
		Mountpoint: "/tmp/var/lib/docker/aufs",
		Opts:       "rw,relatime",
		Fstype:     "ext4",
		Source:     "/dev/sda1",
		VfsOpts:    "rw,errors=remount-ro,data=ordered",
	},
	{
		ID:         152,
		Parent:     142,
		Major:      0,
		Minor:      42,
		Root:       "/",
		Mountpoint: "/tmp/aufs/diff/27fa0900fe22",
		Opts:       "rw,relatime",
		Fstype:     "ext4",
		Source:     "/dev/sda1",
		VfsOpts:    "rw,si=dd417b17e9d4a58b,dio,dirperm1",
	},
	{
		ID:         153,
		Parent:     25,
		Major:      0,
		Minor:      43,
		Root:       "/",
		Mountpoint: "/tmp/var/lib/docker/containers/27fa0900fe22/shm",
		Opts:       "rw,nosuid,nodev,noexec,relatime",
		Optional:   "shared:129",
		Fstype:     "tmpfs",
		Source:     "shm",
		VfsOpts:    "rw,size=65536k",
	},
	{
		ID:         159,
		Parent:     25,
		Major:      0,
		Minor:      44,
		Root:       "/",
		Mountpoint: "/tmp/var/lib/docker/containers/27fa0900fe22/zfs",
		Opts:       "rw,nosuid,nodev,noexec,relatime",
		Optional:   "shared:129",
		Fstype:     "zfs",
		Source:     "snapzfs",
		VfsOpts:    "rw,size=65536k",
	},
}
