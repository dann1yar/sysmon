package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"fyne.io/systray"

	"sysmon/internal/metrics"
)

var (
	interval = flag.Duration("interval", 2*time.Second, "интервал обновления")
	diskPath = flag.String("disk", "/", "путь для измерения занятого места на диске")
)

func main() {
	flag.DurationVar(interval, "i", 2*time.Second, "сокращение для -interval")
	flag.Parse()

	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetIcon(statusIcon(0))
	systray.SetTitle("sysmon")
	systray.SetTooltip("sysmon — системный монитор")

	mCPU := systray.AddMenuItem("CPU: —", "")
	mRAM := systray.AddMenuItem("RAM: —", "")
	mDisk := systray.AddMenuItem("Disk: —", "")
	mNet := systray.AddMenuItem("Net: —", "")
	mTemp := systray.AddMenuItem("Temp: —", "")
	mCPU.Disable()
	mRAM.Disable()
	mDisk.Disable()
	mNet.Disable()
	mTemp.Disable()

	systray.AddSeparator()
	mOpen := systray.AddMenuItem("Открыть live-дашборд", "Запустить sysmon в терминале")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Завершить sysmon-tray")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openDashboard()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-sigCh:
				systray.Quit()
				return
			}
		}
	}()

	go updateLoop(mCPU, mRAM, mDisk, mNet, mTemp)
}

func updateLoop(mCPU, mRAM, mDisk, mNet, mTemp *systray.MenuItem) {
	s1 := metrics.Take()
	for {
		time.Sleep(*interval)
		s2 := metrics.Take()
		secs := interval.Seconds()

		cpuPct := metrics.CPUPercent(s1.CPUBusy, s1.CPUTotal, s2.CPUBusy, s2.CPUTotal)
		mem := metrics.ReadMem()
		disk := metrics.ReadDiskUsage(*diskPath)
		diskR := metrics.Rate(s1.DiskRead, s2.DiskRead, secs)
		diskW := metrics.Rate(s1.DiskWrite, s2.DiskWrite, secs)

		systray.SetIcon(statusIcon(cpuPct))
		systray.SetTitle(fmt.Sprintf("CPU %.0f%% · RAM %.0f%%", cpuPct, mem.Percent()))

		mCPU.SetTitle(fmt.Sprintf("CPU:   %.1f%%", cpuPct))
		mRAM.SetTitle(fmt.Sprintf("RAM:   %.1f%% (%s / %s)", mem.Percent(),
			metrics.HumanBytes(float64(mem.UsedKiB())*1024), metrics.HumanBytes(float64(mem.TotalKiB)*1024)))
		mDisk.SetTitle(fmt.Sprintf("Disk:  %.1f%% (%s / %s)  R: %s  W: %s", disk.Percent(),
			metrics.HumanBytes(float64(disk.UsedBytes)), metrics.HumanBytes(float64(disk.TotalBytes)),
			metrics.HumanRate(diskR), metrics.HumanRate(diskW)))

		if s1.Iface != "" && s1.Iface == s2.Iface {
			mNet.SetTitle(fmt.Sprintf("Net:   %s  ↓ %s  ↑ %s", metrics.DisplayIface(s2.Iface),
				metrics.HumanRate(metrics.Rate(s1.RX, s2.RX, secs)),
				metrics.HumanRate(metrics.Rate(s1.TX, s2.TX, secs))))
		} else {
			mNet.SetTitle("Net:   нет активного соединения")
		}

		if temp, ok := metrics.ReadCPUTemp(); ok {
			mTemp.SetTitle(fmt.Sprintf("Temp:  %.1f°C", temp))
		} else {
			mTemp.SetTitle("Temp:  н/д")
		}

		s1 = s2
	}
}

func openDashboard() {
	sysmonPath := resolveSysmonPath()
	if sysmonPath == "" {
		return
	}

	candidates := []struct {
		cmd  string
		args []string
	}{
		{"gnome-terminal", []string{"--", sysmonPath}},
		{"x-terminal-emulator", []string{"-e", sysmonPath}},
	}
	for _, c := range candidates {
		if _, err := exec.LookPath(c.cmd); err == nil {
			_ = exec.Command(c.cmd, c.args...).Start()
			return
		}
	}
}

func resolveSysmonPath() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "sysmon")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if p, err := exec.LookPath("sysmon"); err == nil {
		return p
	}
	return ""
}
