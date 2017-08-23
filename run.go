package main

import (
	"os"
	"strings"

	"github.com/kasheemlew/xperiMoby/cgroups"
	"github.com/kasheemlew/xperiMoby/cgroups/subsystems"
	"github.com/kasheemlew/xperiMoby/container"
	"github.com/sirupsen/logrus"
)

// Run envokes the command
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		logrus.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}
	cgroupManager := cgroups.NewCgroupManager("xperiMoby-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	// add container processes to cgroup
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray, writePipe)
	parent.Wait()
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	logrus.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
