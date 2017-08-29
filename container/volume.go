package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// NewWorkSpace create container file system
func NewWorkSpace(volume, imageName, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeURLExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
			logrus.Infof("%q", volumeURLs)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}
}

// CreateReadOnlyLayer extract busybox.tar to the busybox directory as read-only layer
func CreateReadOnlyLayer(imageName string) {
	unTarFolderURL := RootURL + imageName + "/"
	imageURL := RootURL + imageName + ".tar"
	exist, err := PathExist(unTarFolderURL)
	if err != nil {
		logrus.Errorf("Fail to judge whether dir %s exists. %v", unTarFolderURL, err)
	}
	if exist == false {
		if err := os.MkdirAll(unTarFolderURL, 0777); err != nil {
			logrus.Errorf("Mkdir %s error. %v", unTarFolderURL, err)
		}
		// func (c *Cmd) CombinedOutput() ([]byte, error)
		// CombinedOutput runs the command and returns its combined standard output and standard error.
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", unTarFolderURL).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error %v", imageURL, err)
		}
	}
}

// CreateWriteLayer create folder `writeLayer` as the only write layer in the container
func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", writeURL, err)
	}
}

// CreateMountPoint create folder `mnt` as mount point
func CreateMountPoint(containerName, imageName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		logrus.Errorf("Mkdir dir %s error. %v", mntURL, err)
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayerURL, containerName)
	tmpImageLocation := RootURL + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	if _, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("%v", err)
		return err
	}
	return nil
}

// MountVolume create & mount volumes
func MountVolume(volumeURLs []string, containerName string) error {
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("Mkdir parent dir %s error. %v", parentURL, err)
	}
	containerURL := volumeURLs[1]
	containerVolumeURL := fmt.Sprintf(MntURL, containerName) + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("Mkdir container dir %s eror. %v", containerURL, err)
	}
	dirs := "dirs=" + parentURL
	if _, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput(); err != nil {
		logrus.Infof("Mount Volume failed. %v", err)
		return err
	}
	return nil
}

// DeleteWorkSpace deletes read and write layer when exit
func DeleteWorkSpace(volume, containerName string) {
	volumeURLs := volumeURLExtract(volume)
	length := len(volumeURLs)
	if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
		DeleteMountPointWithVolume(volumeURLs, containerName)
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

// DeleteMountPointWithVolume unmount volume mount point fs & container mount point fs & remove  the mount points
func DeleteMountPointWithVolume(volumeURLs []string, containerName string) error {
	// unmount volume mount point fs
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerURL := mntURL + volumeURLs[1]
	if _, err := exec.Command("umount", containerURL).CombinedOutput(); err != nil {
		logrus.Errorf("Umount volume failed. %v", err)
		return err
	}
	// unmount container fs mount point
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("Umount mountpoint failed. %v", err)
		return err
	}
	// remove fs mount point
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("Remove monutpoint dir %s error %v", MntURL, err)
		return err
	}
	fileinfo, err := os.Stat(mntURL)
	if err == nil {
		logrus.Infof("after delete %v", fileinfo)
	}
	return nil
}

// DeleteMountPoint delete mount point and remove dir
func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("delete mount point %s error. %v", mntURL, err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("Remove dir %s error. %v", mntURL, err)
		return err
	}

	fileinfo, err := os.Stat(mntURL)
	if err == nil {
		logrus.Infof("after delete %v", fileinfo)
	}
	return nil
}

// DeleteWriteLayer deletes write dir
func DeleteWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("Remove dir %s error %v.", writeURL, err)
		return err
	}
	return nil
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
