package metrics

import (
	"os"
	"strconv"
	"strings"
)

type MemStats struct {
	TotalKiB, AvailKiB uint64
}

func ReadMem() MemStats {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemStats{}
	}
	var m MemStats
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "MemTotal:":
			m.TotalKiB, _ = strconv.ParseUint(fields[1], 10, 64)
		case "MemAvailable:":
			m.AvailKiB, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}
	return m
}

func (m MemStats) UsedKiB() uint64 {
	if m.AvailKiB > m.TotalKiB {
		return 0
	}
	return m.TotalKiB - m.AvailKiB
}

func (m MemStats) Percent() float64 {
	if m.TotalKiB == 0 {
		return 0
	}
	return float64(m.UsedKiB()) / float64(m.TotalKiB) * 100
}
