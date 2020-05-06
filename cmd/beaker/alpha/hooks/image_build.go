package hooks

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

type imageBuild struct{}

func (i *imageBuild) EnableScriptlet() string {
	return "${hook_dir}/image_build"
}

func (i *imageBuild) Render(topDir string, opts *CommitHooks) error {
	f, err := os.OpenFile(filepath.Join(topDir, ".git", "hooks", "image_build"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := map[string]interface{}{
		"Workspace":          opts.workspace,
		"TopDir":             topDir,
		"ProjectName":        path.Base(topDir),
		"RegisterWithBeaker": !opts.imageBuild.registerWithBeaker,
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"sanitize": sanitize,
	}).Parse(imageBuildScript)
	if err != nil {
		return err
	}
	if err = t.Execute(f, templateData); err != nil {
		return fmt.Errorf("generating commit hook file: %w", err)
	}

	return nil
}

const (
	imageBuildScript = `#!/bin/sh
# Commit hook installed by beaker alpha commit-hooks image-build

hash=$(git rev-parse --short HEAD)
pushd {{ .TopDir }} 2>&1 > /dev/null
DOCKER_BUILDKIT=1 docker build . -t {{ .ProjectName }}:$hash
{{ if not .RegisterWithBeaker }}
beaker image create {{ if .Workspace -}} -w {{ .Workspace }} {{ end -}} \
	-n {{ .ProjectName | sanitize }}_$hash \
	{{ .ProjectName }}:$hash
{{ end }}
popd 2>&1 > /dev/null
`
)
