package config

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config is a structured representation of a Beaker config file.
type Config struct {
	// Client settings
	BeakerAddress    string `yaml:"agent_address"` // TODO: Find a better name than "agent_address"
	UserToken        string `yaml:"user_token"`
	DefaultWorkspace string `yaml:"default_workspace"`
}

const (
	addressKey          = "BEAKER_ADDR"
	configPathKey       = "BEAKER_CONFIG"
	configPathKeyLegacy = "BEAKER_CONFIG_FILE" // TODO: Remove when we're sure it's unused.
	tokenKey            = "BEAKER_TOKEN"
	defaultAddress      = "https://beaker.org"
	beakerConfigFile    = "config.yml"
)

var beakerConfigDir = filepath.Join(os.Getenv("HOME"), ".beaker")

// New reads environment and configuration files and returns the resulting Beaker configuration.
func New() (*Config, error) {
	// Set up default config before doing anything.
	config := Config{
		BeakerAddress: defaultAddress,
	}

	err := ReadConfigFromFile(GetFilePath(), &config)
	if err != nil {
		return nil, err
	}

	// Environment variables override config.
	if env, ok := os.LookupEnv(addressKey); ok {
		config.BeakerAddress = env
	}
	if env, ok := os.LookupEnv(tokenKey); ok {
		config.UserToken = env
	}

	return &config, nil
}

func GetFilePath() string {
	// Check the path override first.
	if env, ok := os.LookupEnv(configPathKey); ok {
		return env
	}
	if env, ok := os.LookupEnv(configPathKeyLegacy); ok {
		return env
	}
	return filepath.Join(beakerConfigDir, beakerConfigFile)
}

func ReadConfigFromFile(path string, config *Config) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	d := yaml.NewDecoder(r)
	if err := d.Decode(config); err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to read config")
	}

	return nil
}

func WriteConfig(config *Config, filePath string) error {
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	dirPath, _ := filepath.Split(filePath)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(filePath, bytes, 0644)
}
