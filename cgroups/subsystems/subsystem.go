package subsystems

type ResourceConfig struct {
	MemoryLimit string // memory limit
	CpuShare    string // cpu time-sharing slices
	CpuSet      string // number of cpu cores
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
