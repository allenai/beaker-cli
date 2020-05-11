package alpha

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/cmd/beaker/alpha/hooks"
)

type autobuild struct {
	workspace string
}

func newAutobuild(parent *kingpin.CmdClause) {
	opts := &autobuild{}

	a := parent.Command("auto-build", "Enable/disable autobuild on the cloud for this repo")
	a.Flag("workspace", "workspace to associate image with").Short('w').Required().StringVar(&opts.workspace)
	a.Action(func(c *kingpin.ParseContext) error {
		topDir, err := hooks.GetTopDir()
		if err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(topDir, "Dockerfile")); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no Dockerfile detected: %w", err)
			}
			return fmt.Errorf("checking for Dockerfile: %w", err)
		}

		cfg := fmt.Sprintf(cloudbuildYaml, opts.workspace)
		if err = ioutil.WriteFile(filepath.Join(topDir, "cloudbuild.yaml"), []byte(cfg), 0600); err != nil {
			return err
		}
		return nil
	})
}

const cloudbuildYaml = `steps:
- name: gcr.io/kaniko-project/executor
  args:
  - --destination=gcr.io/$PROJECT_ID/autobuild/$REPO_NAME:$SHORT_SHA
  - --cache=true
- name: gcr.io/cloud-builders/gcloud
  entrypoint: 'bash'
  args: [ '-c', 'echo "user_token: ` + "`gcloud secrets versions access latest --secret=beaker-token`\"" + ` >> beaker-config.yml' ]
- name: 'gcr.io/cloud-builders/curl'
  args: ['-L', '-O', 'https://github.com/allenai/beaker/releases/download/v20200430/beaker_linux.tar.gz']
- name: 'ubuntu'
  args: ['tar', '-xvzf', 'beaker_linux.tar.gz']
- name: 'gcr.io/cloud-builders/docker'
  args: ['pull', 'gcr.io/$PROJECT_ID/autobuild/$REPO_NAME:$SHORT_SHA']
- name: 'gcr.io/cloud-builders/gcloud'
  entrypoint: './beaker'
  args: ['image', 'create', '-w', '%s', '-n', '$REPO_NAME-$SHORT_SHA', 'gcr.io/$PROJECT_ID/autobuild/$REPO_NAME:$SHORT_SHA']
  env:
  - 'BEAKER_CONFIG_FILE=beaker-config.yml'
timeout: 3600s
`
