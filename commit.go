package main

import (
	"fmt"
	"os/exec"

	"github.com/kasheemlew/xperiMoby/container"
	"github.com/sirupsen/logrus"
)

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntURL, containerName)
	imageTar := container.RootURL + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("Tar folder %s error %v", mntURL, err)
	}
}
