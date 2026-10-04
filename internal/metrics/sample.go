package metrics

// percent-based metrics (RAM, disk usage, temperature) aren't here — they
// don't need two samples, so they're read fresh at render time instead
type Sample struct {
	CPUBusy, CPUTotal   uint64
	Iface               string
	RX, TX              uint64
	DiskRead, DiskWrite uint64
}

func Take() Sample {
	var s Sample
	s.CPUBusy, s.CPUTotal = ReadCPU()
	s.Iface = PickIface()
	if s.Iface != "" {
		s.RX, s.TX = ReadNet(s.Iface)
	}
	s.DiskRead, s.DiskWrite = ReadDiskIO()
	return s
}
