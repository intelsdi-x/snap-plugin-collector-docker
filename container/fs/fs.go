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

// Package contains code from Google cAdvisor (https://github.com/google/cadvisor) with following:
// - function collecting disk usage data from running containers

// Package fs provides filesystem statistics
package fs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	zfs "github.com/mistifyio/go-zfs"
	"github.com/moby/moby/pkg/mount"

	"github.com/intelsdi-x/snap-plugin-collector-docker/config"
	"github.com/intelsdi-x/snap-plugin-collector-docker/container"
)

const (
	aufsStorageDriver    = "aufs"
	overlayStorageDriver = "overlay"

	// The read write aufs layers exist here.
	aufsRWLayer = "diff"

	// Path to the directory where docker stores log files if the json logging driver is enabled.
	pathToContainersDir = "containers"

	userLayerFirstVersionMaj = 1
	userLayerFirstVersionMin = 10
	userLayerIDFile          = "mount-id"

	labelSystemRoot = "root"
)

var collector *DiskUsageCollector

// RealFsInfo holds information about filesystem (e.g. partitions)
type RealFsInfo struct {
	// Map from block device path to partition information.
	partitions map[string]partition

	// Map from label to block device path.
	// Labels are intent-specific tags that are auto-detected.
	labels map[string]string
}

// DiskUsageCollector collects container disk usage (based on `du -u` command) in the background
type DiskUsageCollector struct {
	Mut         *sync.Mutex
	DiskUsage   map[string]uint64
	initialized bool
}

type partition struct {
	mountpoint string
	major      uint
	minor      uint
	fsType     string
	blockSize  uint
}

var partitionRegex = regexp.MustCompile(`^(?:(?:s|xv)d[a-z]+\d*|dm-\d+)$`)

func (c *DiskUsageCollector) Init() {
	collector = c
	collector.initialized = true
	collector.DiskUsage = map[string]uint64{}
	collector.Mut = &sync.Mutex{}

	storagePaths := []string{
		"/var/lib/docker",
		"/var/lib/docker/aufs/diff",
		"/var/lib/docker/overlay",
		"/var/lib/docker/zfs",
		"/var/lib/docker/containers",
	}

	collector.worker(false, "root", storagePaths[0])
	collector.worker(true, "containers", storagePaths[1:]...)

}

func (c *DiskUsageCollector) worker(forSubDirs bool, id string, paths ...string) {
	go func(forSubDirs bool, id string, paths ...string) {
		dirs := []string{}
		for _, p := range paths {
			if forSubDirs {
				subdirs, _ := ioutil.ReadDir(p)
				for _, sd := range subdirs {
					dirs = append(dirs, path.Join(p, sd.Name()))
				}
			} else {
				dirs = append(dirs, paths...)
			}
		}

		if len(dirs) > 0 {
			for {
				for _, d := range dirs {
					args := []string{"-sx", d}
					size, err := diskUsage("du", args)
					if err != nil {
						fmt.Fprintf(os.Stderr, "WORKER %s, ERROR {%s} for %s\n", id, err, d)
						break
					}
					c.Mut.Lock()
					c.DiskUsage[d] = size
					c.Mut.Unlock()
				}
				time.Sleep(30 * time.Second)
			}
		} else {
			fmt.Fprintf(os.Stderr, "WORKER %s, ERROR no storage points to collect", id)
		}
	}(forSubDirs, id, paths...)
}

// GetDirFsDevice returns the block device info of the filesystem on which 'dir' resides.
func (self *RealFsInfo) GetDirFsDevice(dir string) (*DeviceInfo, error) {
	buf := new(syscall.Stat_t)
	err := syscall.Stat(dir, buf)

	if err != nil {
		return nil, fmt.Errorf("stat failed on %s with error: %s", dir, err)
	}
	major := major(buf.Dev)
	minor := minor(buf.Dev)
	for device, partition := range self.partitions {
		if partition.major == major && partition.minor == minor {
			return &DeviceInfo{device, major, minor}, nil
		}
	}
	return nil, fmt.Errorf("could not find device with major: %d, minor: %d in cached partitions map", major, minor)
}

// GetDirUsage returns number of bytes occupied by 'dir'.
func (self *RealFsInfo) GetDirUsage(dir string, timeout time.Duration) (uint64, error) {
	collector.Mut.Lock()
	size, ok := collector.DiskUsage[dir]
	collector.Mut.Unlock()
	if !ok {
		return 0, fmt.Errorf("Disk usage not found for %s", dir)
	}
	return size * 1024, nil
}

// GetFsInfoForPath returns capacity and free space, in bytes, of the set of mounts passed.
func (self *RealFsInfo) GetFsInfoForPath(mountSet map[string]struct{}, procfs string) ([]Fs, error) {
	var filesystems []Fs
	deviceSet := make(map[DeviceId]struct{})

	diskStatsMap, err := getDiskStatsMap(filepath.Join(procfs, "diskstats"))
	if err != nil {
		return nil, err
	}

	for device, partition := range self.partitions {
		id := DeviceId{
			Major: uint(partition.major),
			Minor: uint(partition.minor),
		}
		_, hasMount := mountSet[partition.mountpoint]
		_, hasDevice := deviceSet[id]
		if mountSet == nil || (hasMount && !hasDevice) {
			var (
				err error
				fs  Fs
			)
			switch partition.fsType {
			case DeviceMapper.String():
				fs.Capacity, fs.Free, fs.Available, err = getDMStats(device, partition.blockSize)
				fs.Type = DeviceMapper
			case ZFS.String():
				fs.Capacity, fs.Free, fs.Available, err = getZfstats(device)
				fs.Type = ZFS
			default:
				fs.Capacity, fs.Free, fs.Available, fs.Inodes, fs.InodesFree, err = getVfsStats(partition.mountpoint)
				fs.Type = VFS
			}
			if err != nil {
				return nil, err
			}

			deviceSet[id] = struct{}{}

			fs.DeviceInfo = DeviceInfo{
				Device: device,
				Major:  uint(partition.major),
				Minor:  uint(partition.minor),
			}

			if diskStats, exist := diskStatsMap[id]; exist {
				fs.DiskStats = diskStats
				filesystems = append(filesystems, fs)
			}

		}
	}

	return filesystems, nil
}

// GetFsStats returns filesystem statistics for a given container
func (du *DiskUsageCollector) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	var (
		baseUsage           uint64
		logUsage            uint64
		logsFilesStorageDir string
	)

	if collector == nil {
		log.WithFields(log.Fields{
			"block":    "fs",
			"function": "GetStats",
		}).Info("Initializing filesystem collector workers")
		du.Init()
	}

	id, err := opts.GetStringValue("container_id")
	if err != nil {
		return err
	}

	drv, err := opts.GetStringValue("container_drv")
	if err != nil {
		return err
	}

	procfs, err := opts.GetStringValue("procfs")
	if err != nil {
		return err
	}

	rootDir, err := opts.GetStringValue("root_dir")
	if err != nil {
		return err
	}
	rootFsStorageDir := rootDir

	if id != "root" {
		getUserLayerID := func(storageDir, storageDriver, containerID string) (string, error) {
			dockerVersion := config.DockerVersion
			if dockerVersion[0] <= userLayerFirstVersionMaj && dockerVersion[1] < userLayerFirstVersionMin {
				return containerID, nil
			}
			switch storageDriver {
			case aufsStorageDriver:
				fallthrough
			case overlayStorageDriver:
				idFilePath := filepath.Join(storageDir, "image", storageDriver, "layerdb", "mounts", containerID, userLayerIDFile)
				idBytes, err := ioutil.ReadFile(idFilePath)
				if err != nil {
					return "", err
				}
				return string(idBytes), nil
			default:
				return "", fmt.Errorf("Unsupported storage driver; dont know how to determine id of user layer for container %v \n", containerID)
			}
		}
		userLayerID, err := getUserLayerID(rootDir, drv, id)
		if err != nil {
			userLayerID = id
		}

		switch drv {
		case aufsStorageDriver:
			// build the path to docker storage as `/var/lib/docker/aufs/diff/<docker_id>`
			rootFsStorageDir = filepath.Join(rootDir, string(aufsStorageDriver), aufsRWLayer, userLayerID)
		case overlayStorageDriver:
			// build the path to docker storage as `/var/lib/docker/overlay/<docker_id>`
			rootFsStorageDir = filepath.Join(rootDir, string(overlayStorageDriver), userLayerID)
		default:
			return fmt.Errorf("Filesystem stats for storage driver %+s have not been supported yet", drv)
		}

		// Path to the directory where docker stores log files, metadata and configs
		// e.g. /var/lib/docker/container/<docker_id>
		logsFilesStorageDir = filepath.Join(rootDir, pathToContainersDir, id)
	}

	fsInfo, err := newFsInfo(drv)
	if err != nil {
		return err
	}

	if _, err := os.Stat(rootFsStorageDir); err == nil {
		deviceInfo, err := fsInfo.GetDirFsDevice(rootFsStorageDir)
		if err != nil {
			return err
		}

		filesystems, err := fsInfo.GetGlobalFsInfo(procfs)
		if err != nil {
			return fmt.Errorf("Cannot get global filesystem info, err=%v", err)
		}

		baseUsage, err = fsInfo.GetDirUsage(rootFsStorageDir, time.Second)
		if err != nil {
			log.WithFields(log.Fields{
				"block":    "filesystem",
				"function": "GetStats",
			}).Errorf("Cannot get usage for dir=`%s`, err=%s", rootFsStorageDir, err)
		}

		if _, err := os.Stat(logsFilesStorageDir); err == nil {
			logUsage, err = fsInfo.GetDirUsage(logsFilesStorageDir, time.Second)
			if err != nil {
				log.WithFields(log.Fields{
					"block":    "filesystem",
					"function": "GetStats",
				}).Errorf("Cannot get usage for dir=`%s`, err=%s", logsFilesStorageDir, err)
			}

			baseUsage += logUsage
		}

		for _, fs := range filesystems {
			if id == "" {
				deviceInfo.Device = fs.Device
			}

			if fs.Device == deviceInfo.Device {
				fsStats := container.FilesystemInterface{
					Device:          fs.Device,
					Type:            fs.Type.String(),
					Available:       fs.Available,
					Limit:           fs.Capacity,
					Usage:           fs.Capacity - fs.Free,
					BaseUsage:       baseUsage,
					InodesFree:      fs.InodesFree,
					ReadsCompleted:  fs.DiskStats.ReadsCompleted,
					ReadsMerged:     fs.DiskStats.ReadsMerged,
					SectorsRead:     fs.DiskStats.SectorsRead,
					ReadTime:        fs.DiskStats.ReadTime,
					WritesCompleted: fs.DiskStats.WritesCompleted,
					WritesMerged:    fs.DiskStats.WritesMerged,
					SectorsWritten:  fs.DiskStats.SectorsWritten,
					WriteTime:       fs.DiskStats.WriteTime,
					IoInProgress:    fs.DiskStats.IoInProgress,
					IoTime:          fs.DiskStats.IoTime,
					WeightedIoTime:  fs.DiskStats.WeightedIoTime,
				}
				if devName := getDeviceName(fs.Device); len(devName) > 0 {
					stats.Filesystem[devName] = fsStats
				} else {
					log.WithFields(log.Fields{
						"block":    "filesystem",
						"function": "GetStats",
					}).Errorf("Unknown device name")
					stats.Filesystem["unknown"] = fsStats
				}
			}
		}
	} else {
		log.WithFields(log.Fields{
			"block":    "filesystem",
			"function": "GetStats",
		}).Errorf("Os.Stat failed: %v; no fs stats will be available for container %v", err, id)
	}

	return nil
}

//GetGlobalFsInfo returns capacity and free space, in bytes, of all the ext2, ext3, ext4 filesystems on the host.
func (self *RealFsInfo) GetGlobalFsInfo(procfs string) ([]Fs, error) {
	return self.GetFsInfoForPath(nil, procfs)
}

// addSystemRootLabel attempts to determine which device contains the mount for /.
func (self *RealFsInfo) addSystemRootLabel(mounts []*mount.Info) {
	for _, m := range mounts {
		if m.Mountpoint == "/" {
			self.partitions[m.Source] = partition{
				fsType:     m.Fstype,
				mountpoint: m.Mountpoint,
				major:      uint(m.Major),
				minor:      uint(m.Minor),
			}
			self.labels[labelSystemRoot] = m.Source
			return
		}
	}
}

// updateContainerImagesPath compares the mountpoints with possible container image mount points; if a match is found,
// the label is added to the partition.
func (self *RealFsInfo) updateContainerImagesPath(label string, mounts []*mount.Info, containerImagePaths map[string]struct{}) {
	var useMount *mount.Info
	for _, m := range mounts {
		if _, ok := containerImagePaths[m.Mountpoint]; ok {
			if useMount == nil || (len(useMount.Mountpoint) < len(m.Mountpoint)) {
				useMount = m
			}
		}
	}
	if useMount != nil {
		self.partitions[useMount.Source] = partition{
			fsType:     useMount.Fstype,
			mountpoint: useMount.Mountpoint,
			major:      uint(useMount.Major),
			minor:      uint(useMount.Minor),
		}
		self.labels[label] = useMount.Source
	}
}

func diskUsage(cmd string, args []string) (uint64, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return 0, err
	}

	val := strings.Fields(string(out))[0]
	size, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func newFsInfo(storageDriver string) (FsInfo, error) {
	var mounts []*mount.Info
	var err error
	if storageDriver == "test" {
		mounts = mockedMounts
	} else {
		mounts, err = mount.GetMounts()
	}
	if err != nil {
		return nil, err
	}
	fsInfo := &RealFsInfo{
		partitions: make(map[string]partition, 0),
		labels:     make(map[string]string, 0),
	}

	fsInfo.addSystemRootLabel(mounts)

	supportedFsType := map[string]bool{
		// all ext systems are checked through prefix.
		"btrfs": true,
		"xfs":   true,
		"zfs":   true,
	}
	for _, mount := range mounts {
		var Fstype string
		if !strings.HasPrefix(mount.Fstype, "ext") && !supportedFsType[mount.Fstype] {
			continue
		}
		// Avoid bind mounts.
		if _, ok := fsInfo.partitions[mount.Source]; ok {
			continue
		}
		if mount.Fstype == "zfs" {
			Fstype = mount.Fstype
		}
		fsInfo.partitions[mount.Source] = partition{
			fsType:     Fstype,
			mountpoint: mount.Mountpoint,
			major:      uint(mount.Major),
			minor:      uint(mount.Minor),
		}
	}

	return fsInfo, nil
}

func getDiskStatsMap(diskStatsFile string) (map[DeviceId]DiskStats, error) {
	diskStatsMap := make(map[DeviceId]DiskStats)
	file, err := os.Open(diskStatsFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Cannot collect filesystem statistics - file %s is not available", diskStatsFile)
			return diskStatsMap, nil
		}
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		if !partitionRegex.MatchString(words[2]) {
			continue
		}
		// examplary /proc/diskstats content (notice that device name is the third word)
		// 8      50 sdd2 40 0 280 223 7 0 22 108 0 330 330
		wordLength := len(words)
		offset := 3
		var stats = make([]uint64, wordLength-offset)
		if len(stats) < 11 {
			return nil, fmt.Errorf("could not parse all 11 columns of %s", diskStatsFile)
		}
		var error error
		for i := offset; i < wordLength; i++ {
			stats[i-offset], error = strconv.ParseUint(words[i], 10, 64)
			if error != nil {
				return nil, error
			}
		}
		diskStats := DiskStats{
			ReadsCompleted:  stats[0],
			ReadsMerged:     stats[1],
			SectorsRead:     stats[2],
			ReadTime:        stats[3],
			WritesCompleted: stats[4],
			WritesMerged:    stats[5],
			SectorsWritten:  stats[6],
			WriteTime:       stats[7],
			IoInProgress:    stats[8],
			IoTime:          stats[9],
			WeightedIoTime:  stats[10],
		}
		major, err := strconv.ParseUint(words[0], 10, 64)
		if err != nil {
			return nil, err
		}
		minor, err := strconv.ParseUint(words[1], 10, 64)
		if err != nil {
			return nil, err
		}

		diskStatsMap[DeviceId{Major: uint(major), Minor: uint(minor)}] = diskStats
	}

	return diskStatsMap, nil
}

func major(devNumber uint64) uint {
	return uint((devNumber >> 8) & 0xfff)
}

func minor(devNumber uint64) uint {
	return uint((devNumber & 0xff) | ((devNumber >> 12) & 0xfff00))
}

func getVfsStats(path string) (total uint64, free uint64, avail uint64, inodes uint64, inodesFree uint64, err error) {
	var s syscall.Statfs_t
	if err = syscall.Statfs(path, &s); err != nil {
		return 0, 0, 0, 0, 0, err
	}
	total = uint64(s.Frsize) * s.Blocks
	free = uint64(s.Frsize) * s.Bfree
	avail = uint64(s.Frsize) * s.Bavail
	inodes = uint64(s.Files)
	inodesFree = uint64(s.Ffree)
	return total, free, avail, inodes, inodesFree, nil
}

func getDMStats(poolName string, dataBlkSize uint) (uint64, uint64, uint64, error) {
	out, err := exec.Command("dmsetup", "status", poolName).Output()
	if err != nil {
		return 0, 0, 0, err
	}

	used, total, err := parseDMStatus(string(out))
	if err != nil {
		return 0, 0, 0, err
	}

	used *= 512 * uint64(dataBlkSize)
	total *= 512 * uint64(dataBlkSize)
	free := total - used

	return total, free, free, nil
}

func parseDMStatus(dmStatus string) (uint64, uint64, error) {
	dmStatus = strings.Replace(dmStatus, "/", " ", -1)
	dmFields := strings.Fields(dmStatus)

	if len(dmFields) < 8 {
		return 0, 0, fmt.Errorf("Invalid dmsetup status output: %s", dmStatus)
	}

	used, err := strconv.ParseUint(dmFields[6], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	total, err := strconv.ParseUint(dmFields[7], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return used, total, nil
}

// getZfstats returns ZFS mount stats using zfsutils
func getZfstats(poolName string) (uint64, uint64, uint64, error) {
	var dataset *zfs.Dataset
	var err error
	if poolName == "snapzfs" {
		dataset = &zfs.Dataset{Name: poolName}
		dataset.Used = 5000
		dataset.Avail = 2000
		dataset.Usedbydataset = 100
	} else {
		dataset, err = zfs.GetDataset(poolName)
	}
	if err != nil {
		return 0, 0, 0, err
	}

	total := dataset.Used + dataset.Avail + dataset.Usedbydataset

	return total, dataset.Avail, dataset.Avail, nil
}

func getDeviceName(device string) string {
	deviceNs := strings.Split(device, "/")
	return deviceNs[len(deviceNs)-1]
}
