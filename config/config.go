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

	r, err := findConfig()
	if err != nil {
		return nil, err
	}
	if r != nil {
		defer r.Close()

		if err := yaml.NewDecoder(r).Decode(&config); err != nil {
			return nil, errors.Wrap(err, "failed to read config")
		}
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

func ReadConfigFromFile(path string) (*Config, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	config := Config{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&config); err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}

	return &config, nil
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

func findConfig() (io.ReadCloser, error) {
	// Check the path override first.
	if env, ok := os.LookupEnv(configPathKey); ok {
		return os.Open(env)
	}
	if env, ok := os.LookupEnv(configPathKeyLegacy); ok {
		return os.Open(env)
	}

	configPaths := []string{
		beakerConfigDir,
		"/etc/beaker",
	}

	for _, p := range configPaths {
		r, err := os.Open(filepath.Join(p, beakerConfigFile))
		if os.IsNotExist(err) {
			continue
		}
		return r, errors.WithStack(err)
	}

	// No config file found; we'll just use defaults.
	return nil, nil
}
