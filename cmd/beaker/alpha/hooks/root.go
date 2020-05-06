package hooks

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// CommitHooks wraps options used in git hook management.
type CommitHooks struct {
	workspace  string
	imageBuild struct {
		registerWithBeaker bool
	}
	datasetFromSrc struct {
		directory string
	}
	remove bool
}

type hook interface {
	EnableScriptlet() string
	Render(topDir string, opts *CommitHooks) error
}

func NewCommitHooks(app *kingpin.Application, parent *kingpin.CmdClause) {
	opts := &CommitHooks{}
	hooks := parent.Command("commit-hooks", "Manage git commit hooks for managing beaker images/datasets")
	buildImgCmd := hooks.Command("build-image", "Build docker image on commit (and optionally register with beaker)")
	buildImgCmd.Flag("register-with-beaker", "Registering the docker image with beaker").Short('b').BoolVar(&opts.imageBuild.registerWithBeaker)
	buildImgCmd.Flag("remove", "Remove the commit hook").Short('r').BoolVar(&opts.remove)
	buildImgCmd.Flag("workspace", "workspace to associate images and datasets with, uses default workspace if not specified").Short('w').StringVar(&opts.workspace)
	// In theory, we can eliminate the need for this if github URLs could be exposed as datasets to beaker directly.
	// This is still useful for local iteration (i.e. we suggest git commit but don't require a push).
	datasetFromSrcCmd := hooks.Command("dataset-from-src", "Upload code as dataset to beaker")
	datasetFromSrcCmd.Arg("directory", "Upload the directory in repo as dataset on commit. Defaults to the root directory of the repo").StringVar(&opts.datasetFromSrc.directory)
	datasetFromSrcCmd.Flag("remove", "Remove the commit hook").Short('r').BoolVar(&opts.remove)
	datasetFromSrcCmd.Flag("workspace", "workspace to associate images and datasets with, uses default workspace if not specified").Short('w').StringVar(&opts.workspace)
	hooks.PreAction(func(c *kingpin.ParseContext) error {
		// Add automatic help generation for the command group.
		var helpSubcommands []string
		hooks.Command("help", "Show help.").Hidden().Default().PreAction(func(c *kingpin.ParseContext) error {
			fullCommand := append([]string{hooks.Model().Name}, helpSubcommands...)
			app.Usage(fullCommand)
			return nil
		}).Arg("command", "Show help on command.").StringsVar(&helpSubcommands)
		return nil
	})
	processHooks := func(curr hook) error {
		topDir, err := GetTopDir()
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(topDir, "Dockerfile")); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no Dockerfile detected: %w", err)
			}
			return fmt.Errorf("checking for Dockerfile: %w", err)
		}

		if err := exec.Command("docker", "-v").Run(); err != nil {
			return fmt.Errorf("checking for docker installation: %w", err)
		}

		hookDir := filepath.Join(topDir, ".git", "hooks")
		if err = os.MkdirAll(hookDir, 0700); err != nil {
			return err
		}

		currHooks, err := loadHooks(topDir)
		if err != nil {
			return err
		}

		enabled := make(map[hook]bool)
		for _, c := range currHooks {
			enabled[c] = true
		}
		currHooks = append(currHooks, curr)
		if enabled[curr] {
			if opts.remove {
				delete(enabled, curr)
			}
		} else {
			enabled[curr] = true
		}

		for hook, _ := range enabled {
			if err = hook.Render(topDir, opts); err != nil {
				return err
			}
		}

		f, err := os.OpenFile(filepath.Join(hookDir, "post-commit"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
		if err != nil {
			return err
		}
		defer f.Close()

		t, err := template.New("").Parse(postCommitScript)
		if err != nil {
			return err
		}

		var hooks []hook
		for h := range enabled {
			hooks = append(hooks, h)
		}
		if err = t.Execute(f, hooks); err != nil {
			return err
		}

		return nil
	}
	buildImgCmd.Action(func(c *kingpin.ParseContext) error {
		return processHooks(&imageBuild{})
	})
	datasetFromSrcCmd.Action(func(c *kingpin.ParseContext) error {
		return processHooks(&datasetFromSrc{})
	})
}

func loadHooks(topDir string) ([]hook, error) {
	imgBuild := &imageBuild{}
	datasetFromSrc := &datasetFromSrc{}

	scriptletMap := map[string]hook{
		imgBuild.EnableScriptlet():       imgBuild,
		datasetFromSrc.EnableScriptlet(): datasetFromSrc,
	}
	hookFile := filepath.Join(topDir, ".git", "hooks", "post-commit")
	b, err := ioutil.ReadFile(hookFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var hooks []hook
	buf := bytes.NewBuffer(b)
	for {
		l, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if hook, ok := scriptletMap[strings.TrimSpace(l)]; ok {
			hooks = append(hooks, hook)
		}
	}
	return hooks, nil
}

func sanitize(identifier string) (string, error) {
	if len(identifier) == 0 {
		return identifier, nil
	}
	if identifier[0] == '-' {
		identifier = identifier[1:]
	}
	reg, err := regexp.Compile("[^a-zA-Z0-9-_]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(identifier, "_"), nil
}

const (
	postCommitScript = `#!/bin/sh

hook_dir="$(dirname "$0")"

{{- range . }}
{{ .EnableScriptlet }}
{{- end }}
`
)
