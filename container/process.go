package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

// ContainerInfo record information about the container
type ContainerInfo struct {
	Pid         string `json:"pid"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status      string `json:"status"`
}

var (
	RUNNING             = "running"
	STOP                = "stopped"
	EXIT                = "exited"
	DefaultInfoLocation = "/var/run/xperiMoby/%s/"
	ConfigName          = "config.json"
	ContainerLogFile    = "container.log"
)

// NewParentProcess comment
func NewParentProcess(tty bool, volume, containerName string) (*exec.Cmd, *os.File) {
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
	mntURL := "/root/xperi/mnt/"
	rootURL := "/root/xperi/"
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Dir = mntURL
	NewWorkSpace(rootURL, mntURL, volume)
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

// NewWorkSpace create container file system
func NewWorkSpace(rootURL, mntURL, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
	if volume != "" {
		volumeURLs := volumeURLExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mntURL, volumeURLs)
			logrus.Infof("%q", volumeURLs)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}
}

// CreateReadOnlyLayer extract busybox.tar to the busybox directory as read-only layer
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExist(busyboxURL)
	if err != nil {
		logrus.Errorf("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			logrus.Errorf("Mkdir %s error. %v", busyboxURL, err)
		}
		// func (c *Cmd) CombinedOutput() ([]byte, error)
		// CombinedOutput runs the command and returns its combined standard output and standard error.
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error %v", busyboxTarURL, err)
		}
	}
}

// CreateWriteLayer create folder `writeLayer` as the only write layer in the container
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

// CreateMountPoint create folder `mnt` as mount point
func CreateMountPoint(rootURL, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

// DeleteWorkSpace deletes read and write layer when exit
func DeleteWorkSpace(rootURL, mntURL, volume string) {
	volumeURLs := volumeURLExtract(volume)
	length := len(volumeURLs)
	if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
		DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
	} else {
		DeleteMountPoint(mntURL)
	}
	DeleteWriteLayer(rootURL)
}

// DeleteMountPointWithVolume unmount volume mount point fs & container mount point fs & remove  the mount points
func DeleteMountPointWithVolume(rootURL, mntURL string, volumeURLs []string) {
	// unmount volume mount point fs
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount volume failed. %v", err)
	}
	// unmount container fs mount point
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount mountpoint failed. %v", err)
	}
	// remove fs mount point
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("Remove monutpoint dir %s error %v", mntURL, err)
	}
}

// DeleteMountPoint delete mount point and remove dir
func DeleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("delete mount point %s error. %v", mntURL, err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("Remove dir %s error. %v", mntURL, err)
	}
}

// DeleteWriteLayer deletes write dir
func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("Remove dir %s error %v.", writeURL, err)
	}
}

// PathExist checks whether file exists
func PathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func volumeURLExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

// MountVolume create & mount volumes
func MountVolume(rootURL, mntURL string, volumeURLs []string) {
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("Mkdir parent dir %s error. %v", parentURL, err)
	}

	containerURL := volumeURLs[1]
	containerVolumeURL := mntURL + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("Mkdir container dir %s eror. %v", containerURL, err)
	}
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Infof("Mount Volume failed. %v", err)
	}
}
