package subsystems

// ResourceConfig contains the resource limits items
type ResourceConfig struct {
	MemoryLimit string // memory limit
	CPUShare    string // CPU time-sharing slices
	CPUSet      string // number of CPU cores
}

// Subsystem describes methods for subsystem instances
type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	// SubsystemsIns is implements of interface Subsystem
	SubsystemsIns = []Subsystem{
		&CPUSetSubSystem{},
		&MemorySubSystem{},
		&CPUSubSystem{},
	}
)
