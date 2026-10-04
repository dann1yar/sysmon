package metrics

import (
	"os"
	"strings"
)

func PickIface() string {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return ""
	}
	wifi := ""
	for _, e := range entries {
		name := e.Name()
		isEth := strings.HasPrefix(name, "en") || strings.HasPrefix(name, "eth")
		isWifi := strings.HasPrefix(name, "wl") || strings.HasPrefix(name, "w1")
		if !isEth && !isWifi {
			continue
		}
		state, _ := os.ReadFile("/sys/class/net/" + name + "/operstate")
		if strings.TrimSpace(string(state)) != "up" {
			continue
		}
		if isEth {
			return name
		}
		if wifi == "" {
			wifi = name
		}
	}
	return wifi
}

func DisplayIface(name string) string {
	switch {
	case strings.HasPrefix(name, "wl"), strings.HasPrefix(name, "w1"):
		return "wifi"
	case strings.HasPrefix(name, "en"), strings.HasPrefix(name, "eth"):
		return "eth"
	default:
		return name
	}
}

func ReadNet(iface string) (rx, tx uint64) {
	rx = readUintFile("/sys/class/net/" + iface + "/statistics/rx_bytes")
	tx = readUintFile("/sys/class/net/" + iface + "/statistics/tx_bytes")
	return
}
