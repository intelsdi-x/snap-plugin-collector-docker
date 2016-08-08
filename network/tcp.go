// Package network provides network Statistics (included TCP and TCP6 stats)
package network

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-collector-docker/mounts"
	"github.com/intelsdi-x/snap-plugin-collector-docker/wrapper"
)

// TcpStatsFromProc returns TCP statistics (e.g. the number of TCP connections in state `established`, `close`, etc.)
func TcpStatsFromProc(rootFs string, pid int) (wrapper.TcpStat, error) {
	return tcpStatsFromProc(rootFs, pid, "net/tcp")
}

// Tcp6StatsFromProc returns TCP6 statistics (e.g. the number of TCP6 connections in state `established`, `close`, etc.)
func Tcp6StatsFromProc(rootFs string, pid int) (wrapper.TcpStat, error) {
	return tcpStatsFromProc(rootFs, pid, "net/tcp6")
}

func tcpStatsFromProc(rootFs string, pid int, file string) (wrapper.TcpStat, error) {
	tcpStatsFile := filepath.Join(rootFs, mounts.ProcfsMountPoint, strconv.Itoa(pid), file)

	tcpStats, err := scanTcpStats(tcpStatsFile)
	if err != nil {
		return tcpStats, fmt.Errorf("Cannot obtain tcp stats: %v", err)
	}

	return tcpStats, nil
}

func scanTcpStats(tcpStatsFile string) (wrapper.TcpStat, error) {
	var stats wrapper.TcpStat

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

	stats = wrapper.TcpStat{
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
