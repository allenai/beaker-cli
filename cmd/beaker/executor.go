package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
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
	// (optional) Path to the Beaker agent's local storage.
	StoragePath string `yaml:"storagePath"`

	// (optional) Parent directory of session home directories. This can be set
	// to an NFS mount to enable roaming profiles. If unset, sessions mount the
	// invoking user's home directory.
	SessionHome string `yaml:"sessionHome"`
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

	if config.SessionHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("couldn't find home directory: %w", err)
		}
		config.SessionHome = home
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
