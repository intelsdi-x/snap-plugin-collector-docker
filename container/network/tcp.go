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

// Package contains code from Google Cadvisor (https://github.com/google/cadvisor) with following:
// - functions collecting network statistics

// Package network provides network Statistics (included TCP and TCP6 stats)
package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/container"

	log "github.com/sirupsen/logrus"
)

type Tcp struct {
	StatsFile string
}

func (tcp *Tcp) GetStats(stats *container.Statistics, opts container.GetStatOpt) error {
	pid, err := opts.GetIntValue("pid")
	if err != nil {
		return err
	}

	isHost, err := opts.GetBoolValue("is_host")
	if err != nil {
		return err
	}

	procfs, err := opts.GetStringValue("procfs")
	if err != nil {
		return err
	}

	if !isHost {
		path := filepath.Join(procfs, strconv.Itoa(pid), tcp.StatsFile)

		switch tcp.StatsFile {
		case "net/tcp":
			stats.Connection.Tcp, err = tcpStatsFromProc(path)
		case "net/tcp6":
			stats.Connection.Tcp6, err = tcpStatsFromProc(path)
		default:
			log.WithFields(log.Fields{
				"module": "network",
				"block":  "GetStats",
			}).Errorf("Unknown tcp stats file %s", tcp.StatsFile)
			return fmt.Errorf("Unknown tcp stats file %s", tcp.StatsFile)
		}

		if err != nil {
			// only log error message
			log.WithFields(log.Fields{
				"module": "network",
				"block":  "GetStats",
			}).Errorf("Unable to get network stats, pid %d, stats file %s: %s", pid, tcp.StatsFile, err)
		}

	}

	return nil
}

func tcpStatsFromProc(tcpStatsFile string) (container.TcpStat, error) {
	tcpStats, err := scanTcpStats(tcpStatsFile)
	if err != nil {
		return tcpStats, fmt.Errorf("Cannot obtain tcp stats: %v", err)
	}

	return tcpStats, nil
}

func scanTcpStats(tcpStatsFile string) (container.TcpStat, error) {
	var stats container.TcpStat

	data, err := ioutil.ReadFile(tcpStatsFile)
	if err != nil {
		return stats, fmt.Errorf("Cannot open %s: %v", tcpStatsFile, err)
	}

	tcpStateMap := map[string]uint64{
		"01": 0, //ESTABLISHED
		"02": 0, //SYN_SENT
		"03": 0, //SYN_RECV
		"04": 0, //FIN_WAIT1
		"05": 0, //FIN_WAIT2
		"06": 0, //TIME_WAIT
		"07": 0, //CLOSE
		"08": 0, //CLOSE_WAIT
		"09": 0, //LAST_ACK
		"0A": 0, //LISTEN
		"0B": 0, //CLOSING
	}

	reader := strings.NewReader(string(data))
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)

	// Discard header line
	if b := scanner.Scan(); !b {
		return stats, scanner.Err()
	}

	for scanner.Scan() {
		line := scanner.Text()
		state := strings.Fields(line)

		if len(state) < 4 {
			return stats, fmt.Errorf("invalid format of TCP stats file %s: %v", tcpStatsFile, line)
		}

		// TCP state is the 4th field.
		// Format: sl local_address rem_address st tx_queue rx_queue tr tm->when retrnsmt  uid timeout inode
		tcpState := state[3]
		_, ok := tcpStateMap[tcpState]
		if !ok {
			return stats, fmt.Errorf("invalid TCP stats line: %v", line)
		}
		tcpStateMap[tcpState]++
	}

	stats = container.TcpStat{
		Established: tcpStateMap["01"],
		SynSent:     tcpStateMap["02"],
		SynRecv:     tcpStateMap["03"],
		FinWait1:    tcpStateMap["04"],
		FinWait2:    tcpStateMap["05"],
		TimeWait:    tcpStateMap["06"],
		Close:       tcpStateMap["07"],
		CloseWait:   tcpStateMap["08"],
		LastAck:     tcpStateMap["09"],
		Listen:      tcpStateMap["0A"],
		Closing:     tcpStateMap["0B"],
	}

	return stats, nil
}
