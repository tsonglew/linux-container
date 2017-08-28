package main

import (
	"os"
	"os/exec"
	"strings"

	_ "github.com/kasheemlew/xperiMoby/nsenter"
	"github.com/sirupsen/logrus"
)

// EnvExecPid get pid when invoking ExecCommand
const EnvExecPid = "xperiMoby_pid"

// EnvExecCmd get exec command when invoking ExecCommand
const EnvExecCmd = "xperiMoby_cmd"

// ExecContainer enters certain ns
func ExecContainer(containerName string, comArray []string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("Exec container getContainerPidByName %s error %v", containerName, err)
		return
	}
	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(EnvExecPid, pid)
	os.Setenv(EnvExecCmd, cmdStr)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Exec container %s error %v", containerName, err)
	}
}
