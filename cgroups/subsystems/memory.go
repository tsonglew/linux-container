package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// MemorySubSystem is an implement of interface SubSystem
type MemorySubSystem struct {
}

// Set set cgroup limit according to cgroupPath
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true)
	if err == nil {
		if res.MemoryLimit != "" {
			// set cgroup memory_limit and write it to memory.limit_in_bytes
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	}
	return err
}

// Remove delete cgroup according to cgroupPath
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err == nil {
		return os.Remove(subsysCgroupPath)
	}
	return err
}

// Apply add process to cgroup according to cgroupPath
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false)
	if err == nil {
		// wirte pid to task of cgroup in virtual file system
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup fail %v ", err)
		}
		return nil
	}
	return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
}

// Name returns subsystem name
func (s *MemorySubSystem) Name() string {
	return "memory"
}
