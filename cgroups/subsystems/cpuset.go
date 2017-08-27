package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CPUSetSubSystem is an implement of interface SubSystem
type CPUSetSubSystem struct {
}

// Set set cgroup limit according to cgroupPath
func (s *CPUSetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err == nil {
		if res.CPUSet != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "CPUset.CPUs"), []byte(res.CPUSet), 0644); err != nil {
				return fmt.Errorf("set cgroup CPUset fail %v", err)
			}
		}
		return nil
	}
	return err
}

// Remove delete cgroup according to cgroupPath
func (s *CPUSetSubSystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err == nil {
		return os.RemoveAll(subsysCgroupPath)
	}
	return err
}

// Apply add process to cgroup according to cgroupPath
func (s *CPUSetSubSystem) Apply(cgroupPath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail %v", err)
		}
		return nil
	}
	return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
}

// Name returns subsystem name
func (s *CPUSetSubSystem) Name() string {
	return "CPUset"
}
