package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kasheemlew/xperiMoby/container"
	"github.com/sirupsen/logrus"
)

func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	// generate container ID
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-01 13:01:01")
	command := strings.Join(commandArray, " ")
	if containerName == "" {
		containerName = id
	}
	containerInfo := &container.ContainerInfo{
		ID:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("Record container info err %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirURL, 0622); err != nil {
		logrus.Errorf("Mkdir error %s error %v", dirURL, err)
		return "", err
	}
	fileName := dirURL + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		logrus.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	// write jsonified data to `config.json`
	if _, err := file.WriteString(jsonStr); err != nil {
		logrus.Errorf("File write string error %v", err)
		return "", err
	}
	return containerName, nil
}

func deleteContainerInfo(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("Remove dir %s error %v", dirURL, err)
	}
}

func randStringBytes(n int) string {
	letterBytes := "1234567890abcdefghijklmnopqrstuvwxyz"
	letterBytesLen := len(letterBytes)
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(letterBytesLen)]
	}
	return string(b)
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("Read file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		logrus.Errorf("GetContainerInfoByName unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

func getContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var contentInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &contentInfo); err != nil {
		return "", err
	}
	return contentInfo.Pid, nil
}
