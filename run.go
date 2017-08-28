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
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volume, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, volume, containerName)
	if parent == nil {
		logrus.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName)
	if err != nil {
		logrus.Errorf("Record container info error %v", err)
		return
	}
	cgroupManager := cgroups.NewCgroupManager("/root/xperi/xperiMoby-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	// add container processes to cgroup
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
		mntURL := "/root/xperi/mnt/"
		rootURL := "/root/xperi/"
		container.DeleteWorkSpace(rootURL, mntURL, volume)
		deleteContainerInfo(containerName)
	}
	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	logrus.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
