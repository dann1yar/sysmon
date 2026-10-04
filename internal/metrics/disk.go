package metrics

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

const sectorSize = 512

type DiskUsage struct {
	TotalBytes, UsedBytes uint64
}

func ReadDiskUsage(path string) DiskUsage {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return DiskUsage{}
	}
	total := st.Blocks * uint64(st.Bsize)
	free := st.Bavail * uint64(st.Bsize)
	if free > total {
		free = total
	}
	return DiskUsage{TotalBytes: total, UsedBytes: total - free}
}

func (d DiskUsage) Percent() float64 {
	if d.TotalBytes == 0 {
		return 0
	}
	return float64(d.UsedBytes) / float64(d.TotalBytes) * 100
}

func isWholeDisk(name string) bool {
	switch {
	case strings.HasPrefix(name, "loop"), strings.HasPrefix(name, "ram"),
		strings.HasPrefix(name, "dm-"), strings.HasPrefix(name, "sr"):
		return false
	case strings.HasPrefix(name, "nvme"):
		return !strings.Contains(name, "p")
	default:
		last := name[len(name)-1]
		return last < '0' || last > '9'
	}
}

func ReadDiskIO() (readBytes, writtenBytes uint64) {
	data, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		name := fields[2]
		if !isWholeDisk(name) {
			continue
		}
		sectorsRead, _ := strconv.ParseUint(fields[5], 10, 64)
		sectorsWritten, _ := strconv.ParseUint(fields[9], 10, 64)
		readBytes += sectorsRead * sectorSize
		writtenBytes += sectorsWritten * sectorSize
	}
	return
}
