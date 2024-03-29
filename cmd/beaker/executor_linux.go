package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

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
)

var (
	// Path where the Beaker token used by the executor is stored.
	executorTokenPath = path.Join(executorConfigDir, "executor-token")

	// Path where the Systemd configuration file for the executor is stored.
	executorSystemdPath = fmt.Sprintf("/etc/systemd/system/%s.service", executorService)
)

var configTemplate = template.Must(template.New("config").Parse(`
storagePath: {{.StoragePath}}
beaker:
  address: {{.Address}}
  tokenPath: {{.TokenPath}}
  cluster: {{.Cluster}}
resources:
  gpus: {{.GPUs}}`))

type configOpts struct {
	StoragePath string
	Address     string
	TokenPath   string
	Cluster     string
	GPUs        string
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
	cmd.AddCommand(newExecutorConfigureCommand())
	cmd.AddCommand(newExecutorInstallCommand())
	cmd.AddCommand(newExecutorRestartCommand())
	cmd.AddCommand(newExecutorStartCommand())
	cmd.AddCommand(newExecutorStopCommand())
	cmd.AddCommand(newExecutorUninstallCommand())
	cmd.AddCommand(newExecutorUpgradeCommand())
	return cmd
}

func newExecutorConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure <cluster>",
		Short: "Configure the Beaker executor.",
		Args:  cobra.ExactArgs(1),
	}

	var cpuOnly bool
	cmd.Flags().BoolVar(
		&cpuOnly,
		"cpu-only",
		false,
		"Don't try to detect GPUs.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cluster := args[0]
		if _, err := beaker.Cluster(args[0]).Get(ctx); err != nil {
			return err
		}

		if err := os.MkdirAll(executorConfigDir, os.ModePerm); err != nil {
			return err
		}

		if err := os.MkdirAll(executorStoragePath, os.ModePerm); err != nil {
			return err
		}

		if err := ioutil.WriteFile(
			executorTokenPath,
			[]byte(beakerConfig.UserToken),
			0600,
		); err != nil {
			return err
		}

		gpus := "null"
		if cpuOnly {
			gpus = "[]"
		}

		configFile, err := os.Create(executorConfigPath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		return configTemplate.Execute(configFile, configOpts{
			StoragePath: executorStoragePath,
			Address:     beakerConfig.BeakerAddress,
			TokenPath:   executorTokenPath,
			Cluster:     cluster,
			GPUs:        gpus,
		})
	}
	return cmd
}

func newExecutorInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install and start the Beaker executor",
		Long: `Install the Beaker executor, start it, and configure it to run on boot.
Requires access to /etc, /var, and /usr/bin. Also requires access to systemd.`,
		Args: cobra.NoArgs,
	}

	var version string
	cmd.Flags().StringVar(
		&version,
		"version",
		"",
		"Version of the Beaker executor. Commit SHA from allenai/beaker-service. Defaults to the latest version if empty.")

	var validate bool
	cmd.Flags().BoolVar(
		&validate,
		"validate",
		true,
		"Validate the executor installation by waiting for the executor to register a node.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(executorPath); err == nil {
			return fmt.Errorf(`executor is already installed.
Run "upgrade" to install the latest version or run "uninstall" before installing.`)
		}

		if _, err := os.Stat(executorConfigPath); os.IsNotExist(err) {
			return fmt.Errorf(`config file not found.
Run "configure" to generate a configuration file or write configuration to %s manually.`, executorConfigPath)
		} else if err != nil {
			return fmt.Errorf("stat config file: %w", err)
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

		if version == "" {
			version, err = getLatestVersion()
			if err != nil {
				return err
			}
		}
		if err := downloadExecutor(version); err != nil {
			return err
		}

		if err := startExecutor(); err != nil {
			return err
		}

		if !quiet {
			fmt.Println("Executor installed.")
		}
		if !validate {
			return nil
		}
		if !quiet {
			fmt.Println("Waiting for initialization to complete...")
		}
		config, err := getExecutorConfig()
		if err != nil {
			return err
		}
		storagePath := config.StoragePath
		if storagePath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			storagePath = path.Join(home, ".beaker", "storage")
		}
		ready := func(ctx context.Context) (bool, error) {
			out, err := run("sudo", "systemctl", "is-active", executorService)
			if err != nil {
				return false, fmt.Errorf("executor status is not active: %s", out)
			}

			// Check if the executor has registered a node.
			_, err = os.Stat(path.Join(storagePath, "node"))
			if os.IsNotExist(err) {
				return false, nil
			}
			if err != nil {
				return false, fmt.Errorf("stat node file: %w", err)
			}
			return true, nil
		}
		ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute))
		defer cancel()
		if err := await(ctx, "Initializing executor", ready, 0); err != nil {
			return fmt.Errorf("error initializing executor: %w", err)
		}
		if !quiet {
			fmt.Println("Executor is ready to use.")
		}
		return nil
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

			if err := startExecutor(); err != nil {
				return err
			}

			if !quiet {
				fmt.Println("Executor restarted")
			}
			return nil
		},
	}
}

func newExecutorStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the executor",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := startExecutor(); err != nil {
				return err
			}

			if !quiet {
				fmt.Println("Executor started")
			}
			return nil
		},
	}
}

func newExecutorStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the executor and all running jobs",
		Long: `Stop the executor and all running jobs.
To reload executor config without stopping running jobs, use restart.`,
		Args: cobra.NoArgs,
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

			if err := cleanupExecutor(); err != nil {
				return err
			}

			if !quiet {
				fmt.Println("Executor stopped")
			}
			return nil
		},
	}
}

func newExecutorUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the executor and delete all executor data",
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

			// This may fail if the systemd file has already been deleted.
			if err := stopExecutor(); err != nil {
				fmt.Fprintf(os.Stderr, "error stopping executor: %v\n", err)
			}

			// This may fail if the executor binary has already been deleted.
			if err := cleanupExecutor(); err != nil {
				fmt.Fprintf(os.Stderr, "error cleaning up executor: %v\n", err)
			}

			if err := os.RemoveAll(config.StoragePath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorTokenPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorSystemdPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorConfigPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if err := os.Remove(executorPath); err != nil && !os.IsNotExist(err) {
				return err
			}

			if !quiet {
				fmt.Println("Executor uninstalled")
			}
			return nil
		},
	}
}

func newExecutorUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade the executor binary to the latest version",
		Long: `Upgrade the executor binary to the latest version.
To update executor configuration, run uninstall and then install.`,
		Args: cobra.NoArgs,
	}

	var version string
	cmd.Flags().StringVar(
		&version,
		"version",
		"",
		"Version of the Beaker executor. Defaults to the latest version if empty.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := stopExecutor(); err != nil {
			return err
		}

		if version == "" {
			var err error
			version, err = getLatestVersion()
			if err != nil {
				return err
			}
		}
		if err := downloadExecutor(version); err != nil {
			return err
		}

		if err := startExecutor(); err != nil {
			return err
		}

		if !quiet {
			fmt.Println("Executor upgraded")
		}
		return nil
	}
	return cmd
}

func downloadExecutor(version string) error {
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
	if _, err := run("systemctl", "daemon-reload"); err != nil {
		return err
	}

	if _, err := run("systemctl", "enable", executorService); err != nil {
		return err
	}

	if _, err := run("systemctl", "start", executorService); err != nil {
		return err
	}
	return nil
}

func stopExecutor() error {
	if _, err := run("systemctl", "disable", executorService); err != nil {
		return err
	}

	_, err := run("systemctl", "stop", executorService)
	return err
}

// The executor cleanup command removes running containers.
func cleanupExecutor() error {
	cmd := exec.CommandContext(ctx, executorPath, "cleanup")
	cmd.Env = []string{strings.Join([]string{"CONFIG_PATH", executorConfigPath}, "=")}
	_, err := runCmd(cmd)
	return err
}

func runCmd(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running from %q:\n%s\n", strings.Join(cmd.Args, " "), out)
		return nil, err
	}
	return out, nil
}

func run(path string, args ...string) ([]byte, error) {
	return runCmd(exec.CommandContext(ctx, path, args...))
}
