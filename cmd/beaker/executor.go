package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	// Location for storing Beaker configuration.
	executorConfigDir = "/etc/beaker"

	// Path to the node file within the executor's storage directory.
	executorNodePath = "node"
)

var (
	// Path where executor configuration is stored.
	executorConfigPath = path.Join(executorConfigDir, "config.yml")
)

type executorConfig struct {
	StoragePath string `yaml:"storagePath"`
}

// Get the config of the executor running on this machine.
func getExecutorConfig() (*executorConfig, error) {
	configFile, err := ioutil.ReadFile(executorConfigPath)
	if err != nil {
		return nil, err
	}
	expanded := strings.NewReader(os.ExpandEnv(string(configFile)))

	var config executorConfig
	if err := yaml.NewDecoder(expanded).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Get the node ID of the executor running on this machine, if there is one.
func getCurrentNode() (string, error) {
	config, err := getExecutorConfig()
	if err != nil {
		return "", err
	}

	node, err := ioutil.ReadFile(path.Join(config.StoragePath, executorNodePath))
	if err != nil {
		return "", err
	}
	return string(node), nil
}
