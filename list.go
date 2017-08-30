package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/kasheemlew/xperiMoby/container"
	"github.com/sirupsen/logrus"
)

// ListContainers list the infos of all containers
func ListContainers() {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		logrus.Errorf("Read dir %s error %v", dirURL, err)
		return
	}

	var containers []*container.ContainerInfo
	// traversal for container infos
	for _, file := range files {
		if file.Name() == "network" {
			continue
		}
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}

	// write to console
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		logrus.Errorf("Read file %s error %v", configFileDir, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		logrus.Errorf("Json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
