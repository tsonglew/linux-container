package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/kasheemlew/xperiMoby/container"
)

// Run envokes the command
func Run(tty bool, command string) {
	parent := container.NewParentProcess(tty, command)
	if err := parent.Start(); err != nil {
		logrus.Error(err)
	}
	parent.Wait()
	os.Exit(-1)
}
