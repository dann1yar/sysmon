package metrics

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// falls back to hwmon (k10temp, coretemp, ...) since some systems don't
// wire up the ACPI thermal zone driver and expose no thermal_zone at all
func ReadCPUTemp() (celsius float64, ok bool) {
	if v, ok := readThermalZoneTemp(); ok {
		return v, true
	}
	return readHwmonCPUTemp()
}

func readThermalZoneTemp() (float64, bool) {
	entries, err := os.ReadDir("/sys/class/thermal")
	if err != nil {
		return 0, false
	}

	fallback := ""
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "thermal_zone") {
			continue
		}
		base := "/sys/class/thermal/" + name
		zoneType := strings.ToLower(strings.TrimSpace(readFile(base + "/type")))

		if fallback == "" {
			fallback = base
		}
		if strings.Contains(zoneType, "cpu") ||
			strings.Contains(zoneType, "x86_pkg_temp") ||
			strings.Contains(zoneType, "coretemp") {
			if v, ok := readMilliDeg(base + "/temp"); ok {
				return v, true
			}
		}
	}

	if fallback == "" {
		return 0, false
	}
	return readMilliDeg(fallback + "/temp")
}

var cpuHwmonNames = []string{"k10temp", "coretemp", "cpu_thermal", "zenpower"}

// AMD reports "Tctl"/"Tdie", Intel "Package id 0"
var cpuHwmonLabels = []string{"tctl", "tdie", "package id 0"}

func readHwmonCPUTemp() (float64, bool) {
	dirs, err := filepath.Glob("/sys/class/hwmon/hwmon*")
	if err != nil {
		return 0, false
	}

	byName := map[string]string{}
	for _, dir := range dirs {
		byName[strings.ToLower(strings.TrimSpace(readFile(dir+"/name")))] = dir
	}

	for _, want := range cpuHwmonNames {
		dir, found := byName[want]
		if !found {
			continue
		}
		if v, ok := readHwmonBestTemp(dir); ok {
			return v, true
		}
	}
	return 0, false
}

func readHwmonBestTemp(dir string) (float64, bool) {
	inputs, err := filepath.Glob(dir + "/temp*_input")
	if err != nil || len(inputs) == 0 {
		return 0, false
	}

	fallback := ""
	for _, input := range inputs {
		if fallback == "" {
			fallback = input
		}
		labelPath := strings.TrimSuffix(input, "_input") + "_label"
		label := strings.ToLower(strings.TrimSpace(readFile(labelPath)))
		for _, want := range cpuHwmonLabels {
			if strings.Contains(label, want) {
				if v, ok := readMilliDeg(input); ok {
					return v, true
				}
			}
		}
	}
	return readMilliDeg(fallback)
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func readMilliDeg(path string) (float64, bool) {
	v, err := strconv.ParseInt(strings.TrimSpace(readFile(path)), 10, 64)
	if err != nil {
		return 0, false
	}
	return float64(v) / 1000, true
}
