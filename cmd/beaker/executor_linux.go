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

	// Default location for storing Beaker configuration.
	defaultConfigDir = "/etc/beaker"

	// Default location for storing datasets.
	defaultStorageDir = "/var/beaker"
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

func newExecutorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executor <command>",
		Short: "Manage the executor",
	}
	cmd.AddCommand(newExecutorInstallCommand())
	return cmd
}

func newExecutorInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <cluster>",
		Short: "Install the Beaker executor",
		Args:  cobra.ExactArgs(1),
	}

	var configDir string
	var storageDir string
	cmd.Flags().StringVar(
		&configDir,
		"config-dir",
		defaultConfigDir,
		"Writeable directory for Beaker configuration")
	cmd.Flags().StringVar(
		&storageDir,
		"storage-dir",
		defaultStorageDir,
		"Writeable directory for storing Beaker datasets")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
			return err
		}

		tokenPath := path.Join(configDir, "token")
		if err := ioutil.WriteFile(
			path.Join(configDir, "token"),
			[]byte(beakerConfig.UserToken),
			0600,
		); err != nil {
			return err
		}

		configPath := path.Join(configDir, "config.yml")
		configFile, err := os.Create(configPath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		if err := configTemplate.Execute(configFile, configOpts{
			StoragePath: storageDir,
			TokenPath:   tokenPath,
			Cluster:     args[0],
		}); err != nil {
			return err
		}

		systemdPath := fmt.Sprintf("/etc/systemd/system/%s.service", executorService)
		systemdFile, err := os.Create(systemdPath)
		if err != nil {
			return err
		}
		defer systemdFile.Close()
		if err := systemdTemplate.Execute(systemdFile, systemdOpts{
			BinaryPath: executorPath,
			ConfigPath: configPath,
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
