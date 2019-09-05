package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Config is a structured representation of a Beaker config file.
type Config struct {
	// Client settings
	BeakerAddress string `yaml:"agent_address"` // TODO: Find a better name than "agent_address"
	UserToken     string `yaml:"user_token"`
	DefaultOrg    string `yaml:"default_org"`
	DefaultWorkspace string `yaml:"default_workspace"`
}

const (
	addressKey     = "BEAKER_ADDR"
	configPathKey  = "BEAKER_CONFIG_FILE"
	defaultAddress = "https://beaker.allenai.org"
)

var beakerConfigDir = filepath.Join(os.Getenv("HOME"), ".beaker")

// New reads environment and configuration files and returns the resulting Beaker configuration.
func New() (*Config, error) {
	// Set up default config before doing anything.
	config := Config{
		BeakerAddress: defaultAddress,
	}
	if addr, ok := os.LookupEnv(addressKey); ok {
		config.BeakerAddress = addr
	}

	r, err := findConfig()
	if err != nil {
		return nil, err
	}
	if r != nil {
		defer r.Close()

		d := yaml.NewDecoder(r)
		if err := d.Decode(&config); err != nil {
			return nil, errors.Wrap(err, "failed to read config")
		}
	}
	return &config, nil
}

func findConfig() (io.ReadCloser, error) {
	// Check the override first.
	if override, ok := os.LookupEnv(configPathKey); ok {
		return os.Open(override)
	}

	const configFile = "config.yml"
	configPaths := []string{
		beakerConfigDir,
		"/etc/beaker",
	}

	for _, p := range configPaths {
		r, err := os.Open(filepath.Join(p, configFile))
		if os.IsNotExist(err) {
			continue
		}
		return r, errors.WithStack(err)
	}

	// No config file found; we'll just use defaults.
	return nil, nil
}
