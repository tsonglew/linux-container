package subsystems

type ResourceConfig struct {
	MemoryLimit string // memory limit
	CpuShare    string // time-sharing timeslice weights
	CpuSet      string // number of cpu cores
}
