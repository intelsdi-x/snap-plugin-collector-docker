/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation
Copyright 2014 Google Inc. All Rights Reserved.

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

// Package contains code from Google cAdvisor (https://github.com/google/cadvisor ) with following:
// - structures holding information about disk usage

package fs

import "time"

const (
	ZFS          FsType = "zfs"
	DeviceMapper FsType = "devicemapper"
	VFS          FsType = "vfs"
)

// DeviceInfo holds device name and major and minor numbers
type DeviceInfo struct {
	Device string
	Major  uint
	Minor  uint
}

// FsType is a docker filesystem type, supported: zfs, vfs and devicemapper
type FsType string

func (ft FsType) String() string {
	return string(ft)
}

// Fs holds information about device (name, minor, major), type, capacity, etc.
type Fs struct {
	DeviceInfo
	Type       FsType
	Capacity   uint64
	Free       uint64
	Available  uint64
	Inodes     uint64
	InodesFree uint64
	DiskStats  DiskStats
}

// DiskStats holds disk statistics
type DiskStats struct {
	ReadsCompleted  uint64
	ReadsMerged     uint64
	SectorsRead     uint64
	ReadTime        uint64
	WritesCompleted uint64
	WritesMerged    uint64
	SectorsWritten  uint64
	WriteTime       uint64
	IoInProgress    uint64
	IoTime          uint64
	WeightedIoTime  uint64
}

// FsInfo specifies methods to get filesystem information and statistics
type FsInfo interface {
	// Returns capacity and free space, in bytes, of all the ext2, ext3, ext4 filesystems on the host.
	GetGlobalFsInfo(procfs string) ([]Fs, error)

	// Returns capacity and free space, in bytes, of the set of mounts passed.
	GetFsInfoForPath(mountSet map[string]struct{}, procfs string) ([]Fs, error)

	// Returns number of bytes occupied by 'dir'.
	GetDirUsage(dir string, timeout time.Duration) (uint64, error)

	// Returns the block device info of the filesystem on which 'dir' resides.
	GetDirFsDevice(dir string) (*DeviceInfo, error)
}
