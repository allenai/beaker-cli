package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	// The version URL must respond to a GET request with the latest version of the executor.
	versionURL = "https://storage.googleapis.com/ai2-beaker-public/bin/latest"

	// Replace %s with the version from the URL above.
	executorURL = "https://storage.googleapis.com/ai2-beaker-public/bin/%s/executor"

	// Path to the executor binary.
	executorPath = "/usr/bin/beaker-executor"

	// Name of the executor's systemd service.
	executorService = "beaker-executor"

	// Location for storing Beaker configuration.
	executorConfigDir = "/etc/beaker"

	// Default location for storing datasets.
	defaultStorageDir = "/var/beaker"

	// Path to the node file within the executor's storage directory.
	executorNodePath = "node"
)

var (
	// Path where executor configuration is stored.
	executorConfigPath = path.Join(executorConfigDir, "config.yml")

	// Path where the Beaker token used by the executor is stored.
	executorTokenPath = path.Join(executorConfigDir, "token")

	// Path where the Systemd configuration file for the executor is stored.
	executorSystemdPath = fmt.Sprintf("/etc/systemd/system/%s.service", executorService)
)

var configTemplate = template.Must(template.New("config").Parse(`
storagePath: {{.StoragePath}}
beaker:
  tokenPath: {{.TokenPath}}
  cluster: {{.Cluster}}`))

type configOpts struct {
	StoragePath string
	TokenPath   string
	Cluster     string
}

var systemdTemplate = template.Must(template.New("systemd").Parse(`
[Unit]
Description=Beaker executor
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
ExecStart={{.BinaryPath}}
Environment=CONFIG_PATH={{.ConfigPath}}

[Install]
WantedBy=multi-user.target`))

type systemdOpts struct {
	BinaryPath string
	ConfigPath string
}

type executorConfig struct {
	StoragePath string `yaml:"storagePath"`
}

func newExecutorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executor <command>",
		Short: "Manage the executor",
	}
	cmd.AddCommand(newExecutorInstallCommand())
	cmd.AddCommand(newExecutorRestartCommand())
	cmd.AddCommand(newExecutorStartCommand())
	cmd.AddCommand(newExecutorStopCommand())
	cmd.AddCommand(newExecutorUninstallCommand())
	cmd.AddCommand(newExecutorUpgradeCommand())
	return cmd
}

func newExecutorInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <cluster>",
		Short: "Install and start the Beaker executor. May require sudo.",
		Long: `Install the Beaker executor, start it, and configure it to run on boot.
Requires access to /etc, /var, and /usr/bin. Also requires access to systemd.`,
		Args: cobra.ExactArgs(1),
	}

	var storageDir string
	cmd.Flags().StringVar(
		&storageDir,
		"storage-dir",
		defaultStorageDir,
		"Writeable directory for storing Beaker datasets")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(executorConfigDir, os.ModePerm); err != nil {
			return err
		}

		if err := ioutil.WriteFile(
			executorTokenPath,
			[]byte(beakerConfig.UserToken),
			0600,
		); err != nil {
			return err
		}

		configFile, err := os.Create(executorConfigPath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		if err := configTemplate.Execute(configFile, configOpts{
			StoragePath: storageDir,
			TokenPath:   executorTokenPath,
			Cluster:     args[0],
		}); err != nil {
			return err
		}

		systemdFile, err := os.Create(executorSystemdPath)
		if err != nil {
			return err
		}
		defer systemdFile.Close()
		if err := systemdTemplate.Execute(systemdFile, systemdOpts{
			BinaryPath: executorPath,
			ConfigPath: executorConfigPath,
		}); err != nil {
			return err
		}

		if err := downloadExecutor(); err != nil {
			return err
		}

		return startExecutor()
	}
	return cmd
}

func newExecutorRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the executor without stopping running jobs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := stopExecutor(); err != nil {
				return err
			}

			return startExecutor()
		},
	}
}

func newExecutorStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the executor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startExecutor()
		},
	}
}

func newExecutorStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the executor and kill all running containers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			confirmed, err := confirm(`Stopping the executor will kill all running tasks.
Are you sure you want to stop the executor?`)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}

			if err := stopExecutor(); err != nil {
				return err
			}

			// The executor cleanup command removes running containers.
			return exec.CommandContext(ctx, executorPath, "cleanup").Run()
		},
	}
}

func newExecutorUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the executor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := getExecutorConfig()
			if err != nil {
				return err
			}

			const prompt = `Uninstalling the executor will kill all running tasks
and delete all data in %q.

Are you sure you want to uninstall the executor?`
			confirmed, err := confirm(fmt.Sprintf(prompt, config.StoragePath))
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}

			if err := stopExecutor(); err != nil {
				return err
			}

			// The executor cleanup command removes running containers.
			/*
				if err := exec.CommandContext(ctx, executorPath, "cleanup").Run(); err != nil {
					return err
				}
			*/

			if err := os.RemoveAll(config.StoragePath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorConfigPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorTokenPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorSystemdPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorPath); err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		},
	}
}

func newExecutorUpgradeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the executor to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := stopExecutor(); err != nil {
				return err
			}

			if err := downloadExecutor(); err != nil {
				return err
			}

			return startExecutor()
		},
	}
}

func downloadExecutor() error {
	version, err := getLatestVersion()
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf(executorURL, version))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(executorPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return os.Chmod(executorPath, 0700)
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	version, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(version)), nil
}

func startExecutor() error {
	if err := exec.CommandContext(ctx, "systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	if err := exec.CommandContext(ctx, "systemctl", "enable", executorService).Run(); err != nil {
		return err
	}

	return exec.CommandContext(ctx, "systemctl", "start", executorService).Run()
}

func stopExecutor() error {
	if err := exec.CommandContext(ctx, "systemctl", "disable", executorService).Run(); err != nil {
		return err
	}

	return exec.CommandContext(ctx, "systemctl", "stop", executorService).Run()
}

// Get the node ID of the executor running on this machine.
func getExecutorConfig() (*executorConfig, error) {
	configFile, err := os.Open(executorConfigPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config executorConfig
	if err := yaml.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
