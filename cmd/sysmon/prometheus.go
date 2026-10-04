package main

import (
	"fmt"
	"io"
)

type promMetric struct {
	name  string
	help  string
	value float64
}

// https://prometheus.io/docs/instrumenting/exposition_formats/
func writePrometheus(w io.Writer, snap Snapshot) {
	metrics := []promMetric{
		{"sysmon_cpu_percent", "Current CPU utilization, percent.", snap.CPUPercent},
		{"sysmon_ram_percent", "Current RAM utilization, percent.", snap.RAMPercent},
		{"sysmon_ram_used_bytes", "RAM currently in use, bytes.", float64(snap.RAMUsedBytes)},
		{"sysmon_ram_total_bytes", "Total installed RAM, bytes.", float64(snap.RAMTotalBytes)},
		{"sysmon_disk_percent", "Disk space used, percent.", snap.DiskPercent},
		{"sysmon_disk_used_bytes", "Disk space used, bytes.", float64(snap.DiskUsedBytes)},
		{"sysmon_disk_total_bytes", "Total disk space, bytes.", float64(snap.DiskTotalBytes)},
		{"sysmon_disk_read_bytes_per_second", "Disk read rate, bytes/s.", snap.DiskReadBps},
		{"sysmon_disk_write_bytes_per_second", "Disk write rate, bytes/s.", snap.DiskWriteBps},
		{"sysmon_net_rx_bytes_per_second", "Network receive rate, bytes/s.", snap.NetRxBps},
		{"sysmon_net_tx_bytes_per_second", "Network transmit rate, bytes/s.", snap.NetTxBps},
	}
	if snap.TempOK {
		metrics = append(metrics, promMetric{"sysmon_temp_celsius", "CPU temperature, Celsius.", snap.TempC})
	}

	for _, m := range metrics {
		fmt.Fprintf(w, "# HELP %s %s\n", m.name, m.help)
		fmt.Fprintf(w, "# TYPE %s gauge\n", m.name)
		fmt.Fprintf(w, "%s %v\n", m.name, m.value)
	}
}
