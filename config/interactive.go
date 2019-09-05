package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// InteractiveConfiguration asks the user for values, using the defaults, then saves.
func InteractiveConfiguration() error {
	fmt.Println("Beaker Configuration")
	fmt.Println("")
	fmt.Println("Press enter to keep the current value of any setting.")
	fmt.Printf("Results will be saved to %v\n\n", color.BlueString(BeakerConfigDir))

	// Create a default config by reading in whatever config currently exists.
	config, err := New()
	if err != nil {
		return err
	}

	if config.BeakerAddress, err = promptValue("Beaker address", config.BeakerAddress); err != nil {
		return err
	}
	if config.UserToken, err = promptValue("User token", config.UserToken); err != nil {
		return err
	}
	if config.DefaultOrg, err = promptValue("Default organization", config.DefaultOrg); err != nil {
		return err
	}

	config.UserToken = strings.TrimSpace(config.UserToken)
	if config.UserToken == "" {
		return errors.New("blank token not allowed")
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(BeakerConfigDir, os.ModePerm); err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(filepath.Join(BeakerConfigDir, "config.yml"), bytes, 0644)
}

func promptValue(name, defaultVal string) (string, error) {
	fmt.Printf("%s [%v]: ", name, defaultVal)

	// Read in a new value.
	reader := bufio.NewReader(os.Stdin)
	newValue, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.WithStack(err)
	}

	// Trim excess space. If no value was specified, use the existing or default value.
	newValue = strings.TrimSpace(newValue)
	if newValue == "" {
		return defaultVal, nil
	}

	return newValue, nil
}
