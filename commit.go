package main

import (
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func commitContainer(imageName string) {
	mntURL := "/root/xperi/mnt"
	imageTar := "/root/xperi/" + imageName + ".tar"
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
