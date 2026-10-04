package main

import (
	"fmt"
	"strings"
	"time"

	"sysmon/internal/metrics"
)

const barWidth = 30

func bar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct / 100 * float64(width))
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", width-filled) + "]"
}

func clearScreen() { fmt.Print("\x1b[H\x1b[2J") }
func hideCursor()  { fmt.Print("\x1b[?25l") }
func showCursor()  { fmt.Print("\x1b[?25h") }

type frame struct {
	cpuPct float64
	mem    metrics.MemStats
	disk   metrics.DiskUsage
	diskR  float64 // bytes/s
	diskW  float64 // bytes/s
	tempC  float64
	tempOK bool

	iface   string
	netDown float64 // bytes/s
	netUp   float64 // bytes/s
}

func render(f frame) {
	clearScreen()
	fmt.Printf("sysmon — %s (Ctrl+C для выхода)\n\n", time.Now().Format("15:04:05"))

	fmt.Printf("CPU   %s %5.1f%%\n", bar(f.cpuPct, barWidth), f.cpuPct)

	memPct := f.mem.Percent()
	fmt.Printf("RAM   %s %5.1f%%  (%s / %s)\n", bar(memPct, barWidth), memPct,
		metrics.HumanBytes(float64(f.mem.UsedKiB())*1024), metrics.HumanBytes(float64(f.mem.TotalKiB)*1024))

	diskPct := f.disk.Percent()
	fmt.Printf("DISK  %s %5.1f%%  (%s / %s)  R: %s  W: %s\n", bar(diskPct, barWidth), diskPct,
		metrics.HumanBytes(float64(f.disk.UsedBytes)), metrics.HumanBytes(float64(f.disk.TotalBytes)),
		metrics.HumanRate(f.diskR), metrics.HumanRate(f.diskW))

	if f.tempOK {
		fmt.Printf("TEMP  %.1f°C\n", f.tempC)
	} else {
		fmt.Printf("TEMP  н/д\n")
	}

	if f.iface != "" {
		fmt.Printf("NET   %-8s ↓ %s   ↑ %s\n", metrics.DisplayIface(f.iface), metrics.HumanRate(f.netDown), metrics.HumanRate(f.netUp))
	} else {
		fmt.Printf("NET   нет активного соединения\n")
	}
}
