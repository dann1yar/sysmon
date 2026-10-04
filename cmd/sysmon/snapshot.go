package main

import (
	"math"
	"time"

	"sysmon/internal/metrics"
)

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

type Snapshot struct {
	CPUPercent float64 `json:"cpu_percent"`

	RAMPercent    float64 `json:"ram_percent"`
	RAMUsedBytes  uint64  `json:"ram_used_bytes"`
	RAMTotalBytes uint64  `json:"ram_total_bytes"`

	DiskPercent    float64 `json:"disk_percent"`
	DiskUsedBytes  uint64  `json:"disk_used_bytes"`
	DiskTotalBytes uint64  `json:"disk_total_bytes"`
	DiskReadBps    float64 `json:"disk_read_bytes_per_sec"`
	DiskWriteBps   float64 `json:"disk_write_bytes_per_sec"`

	TempC  float64 `json:"temp_c,omitempty"`
	TempOK bool    `json:"temp_ok"`

	NetIface string  `json:"net_iface,omitempty"`
	NetRxBps float64 `json:"net_rx_bytes_per_sec"`
	NetTxBps float64 `json:"net_tx_bytes_per_sec"`
}

func takeSnapshot(diskPath string, interval time.Duration) Snapshot {
	s1 := metrics.Take()
	time.Sleep(interval)
	s2 := metrics.Take()
	return buildSnapshot(s1, s2, interval.Seconds(), diskPath)
}

func buildSnapshot(s1, s2 metrics.Sample, secs float64, diskPath string) Snapshot {
	mem := metrics.ReadMem()
	disk := metrics.ReadDiskUsage(diskPath)

	snap := Snapshot{
		CPUPercent: round1(metrics.CPUPercent(s1.CPUBusy, s1.CPUTotal, s2.CPUBusy, s2.CPUTotal)),

		RAMPercent:    round1(mem.Percent()),
		RAMUsedBytes:  mem.UsedKiB() * 1024,
		RAMTotalBytes: mem.TotalKiB * 1024,

		DiskPercent:    round1(disk.Percent()),
		DiskUsedBytes:  disk.UsedBytes,
		DiskTotalBytes: disk.TotalBytes,
		DiskReadBps:    metrics.Rate(s1.DiskRead, s2.DiskRead, secs),
		DiskWriteBps:   metrics.Rate(s1.DiskWrite, s2.DiskWrite, secs),
	}

	snap.TempC, snap.TempOK = metrics.ReadCPUTemp()
	snap.TempC = round1(snap.TempC)

	if s1.Iface != "" && s1.Iface == s2.Iface {
		snap.NetIface = s2.Iface
		snap.NetRxBps = metrics.Rate(s1.RX, s2.RX, secs)
		snap.NetTxBps = metrics.Rate(s1.TX, s2.TX, secs)
	}
	return snap
}
