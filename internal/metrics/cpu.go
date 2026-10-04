package metrics

import (
	"os"
	"strconv"
	"strings"
)

func ReadCPU() (busy, total uint64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	line, _, _ := strings.Cut(string(data), "\n")
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, 0
	}

	for i, f := range fields[1:] {
		v, _ := strconv.ParseUint(f, 10, 64)
		total += v
		// i is offset by one vs the "cpu" field: idx 3 = idle, idx 4 = iowait
		if i != 3 && i != 4 {
			busy += v
		}
	}
	return busy, total
}

func CPUPercent(b1, t1, b2, t2 uint64) float64 {
	dt := t2 - t1
	if dt == 0 || t2 < t1 {
		return 0
	}
	return float64(b2-b1) / float64(dt) * 100
}
