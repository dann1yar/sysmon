package metrics

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func readUintFile(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	return v
}

func HumanRate(bps float64) string {
	return HumanBytes(bps) + "/s"
}

func HumanBytes(b float64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", b, units[i])
}

// guards against counter resets, e.g. interface replaced or device unplugged
func Rate(v1, v2 uint64, secs float64) float64 {
	if v2 < v1 || secs <= 0 {
		return 0
	}
	return float64(v2-v1) / secs
}
