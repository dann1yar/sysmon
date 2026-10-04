package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sysmon/internal/metrics"
)

func main() {
	interval := flag.Duration("interval", time.Second, "интервал обновления")
	once := flag.Bool("once", false, "снять один замер и выйти (без живого дашборда)")
	diskPath := flag.String("disk", "/", "путь для измерения занятого места на диске")
	jsonOut := flag.Bool("json", false, "печатать снятый замер в JSON (требует -once)")
	serve := flag.Bool("serve", false, "поднять HTTP-сервер с /metrics (Prometheus) и /json")
	addr := flag.String("addr", "127.0.0.1:9101", "адрес:порт для -serve")
	flag.DurationVar(interval, "i", time.Second, "сокращение для -interval")
	flag.BoolVar(once, "o", false, "сокращение для -once")
	flag.Parse()

	if *interval <= 0 {
		fmt.Fprintln(os.Stderr, "interval должен быть положительным")
		os.Exit(1)
	}
	if *jsonOut && !*once {
		fmt.Fprintln(os.Stderr, "-json можно использовать только вместе с -once")
		os.Exit(1)
	}
	if *serve && (*once || *jsonOut) {
		fmt.Fprintln(os.Stderr, "-serve нельзя сочетать с -once/-json")
		os.Exit(1)
	}

	switch {
	case *once:
		runOnce(*diskPath, *interval, *jsonOut)
	case *serve:
		runServe(*addr, *diskPath, *interval)
	default:
		runLive(*diskPath, *interval)
	}
}

func runOnce(diskPath string, interval time.Duration, jsonOut bool) {
	snap := takeSnapshot(diskPath, interval)
	if jsonOut {
		printSnapshotJSON(snap)
		return
	}
	printSnapshotText(snap)
}

func printSnapshotText(snap Snapshot) {
	fmt.Printf("CPU: %.1f%%\n", snap.CPUPercent)

	fmt.Printf("RAM: %.1f%% (%s / %s)\n", snap.RAMPercent,
		metrics.HumanBytes(float64(snap.RAMUsedBytes)), metrics.HumanBytes(float64(snap.RAMTotalBytes)))

	fmt.Printf("DISK: %.1f%% (%s / %s)  R: %s  W: %s\n", snap.DiskPercent,
		metrics.HumanBytes(float64(snap.DiskUsedBytes)), metrics.HumanBytes(float64(snap.DiskTotalBytes)),
		metrics.HumanRate(snap.DiskReadBps), metrics.HumanRate(snap.DiskWriteBps))

	if snap.TempOK {
		fmt.Printf("TEMP: %.1f°C\n", snap.TempC)
	} else {
		fmt.Println("TEMP: н/д")
	}

	if snap.NetIface != "" {
		fmt.Printf("NET %s: ↓ %s  ↑ %s\n", snap.NetIface,
			metrics.HumanRate(snap.NetRxBps), metrics.HumanRate(snap.NetTxBps))
	} else {
		fmt.Println("NET: нет активного соединения")
	}
}

func runLive(diskPath string, interval time.Duration) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	hideCursor()
	defer showCursor()

	s1 := metrics.Take()
	for {
		select {
		case <-sigCh:	
			return
		case <-time.After(interval):
		}

		s2 := metrics.Take()
		secs := interval.Seconds()

		f := frame{
			cpuPct: metrics.CPUPercent(s1.CPUBusy, s1.CPUTotal, s2.CPUBusy, s2.CPUTotal),
			mem:    metrics.ReadMem(),
			disk:   metrics.ReadDiskUsage(diskPath),
			diskR:  metrics.Rate(s1.DiskRead, s2.DiskRead, secs),
			diskW:  metrics.Rate(s1.DiskWrite, s2.DiskWrite, secs),
			iface:  s2.Iface,
		}
		if s1.Iface != "" && s1.Iface == s2.Iface {
			f.netDown = metrics.Rate(s1.RX, s2.RX, secs)
			f.netUp = metrics.Rate(s1.TX, s2.TX, secs)
		}
		f.tempC, f.tempOK = metrics.ReadCPUTemp()

		render(f)
		s1 = s2
	}
}