package subsystems

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// memory subsystem
type MemorySubSystem struct {
}

// set cgroup limit according to cgroupPath
func (s *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			// set cgroup memory_limit and write it to memory.limit_in_bytes
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit()), 0644); err != nil {
				return fmt.Error("set cgroup memory fail %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

// delete cgroup according to cgroupPath
func (s *MemorySubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// remove the directory
		return os.Remove(subsysCgroupPath)
	} else {
		return err
	}
}

// add process to cgroup according to cgroupPath
func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		// wirte pid to task of cgroup in virtual file system
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			fmt.Errorf("set cgroup fail %v ", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

func (s *MemorySubSystem) Name() string {
	return "memory"
}
