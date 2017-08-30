package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// ContainerInfo record information about the container
type ContainerInfo struct {
	Pid         string   `json:"pid"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreatedTime string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portmapping"`
}

var (
	RUNNING             = "running"
	STOP                = "stopped"
	EXIT                = "exited"
	MntURL              = "/root/xperi/mnt/%s/"
	RootURL             = "/root/xperi/"
	WriteLayerURL       = "/root/xperi/writeLayer/%s/"
	DefaultInfoLocation = "/var/run/xperiMoby/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)

// NewParentProcess comment
func NewParentProcess(tty bool, volume, containerName, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// generate `container.log` & write log to it
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	cmd.Dir = fmt.Sprintf(MntURL, containerName)
	NewWorkSpace(volume, imageName, containerName)
	return cmd, writePipe
}

// NewPipe create read and write pipes
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
