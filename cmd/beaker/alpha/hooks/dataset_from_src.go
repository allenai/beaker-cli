package hooks

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

type datasetFromSrc struct{}

func (d *datasetFromSrc) EnableScriptlet() string {
	return "${hook_dir}/dataset_from_src"
}

func (d *datasetFromSrc) Render(topDir string, opts *CommitHooks) error {
	f, err := os.OpenFile(filepath.Join(topDir, ".git", "hooks", "dataset_from_src"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer f.Close()

	directory := opts.datasetFromSrc.directory
	if directory == "" {
		directory = topDir
	}
	templateData := map[string]interface{}{
		"Workspace":   opts.workspace,
		"TopDir":      topDir,
		"ProjectName": path.Base(topDir),
		"Directory":   directory,
	}

	t, err := template.New("").Funcs(template.FuncMap{
		"sanitize": sanitize,
	}).Parse(datasetCreateScript)
	if err != nil {
		return err
	}
	if err = t.Execute(f, templateData); err != nil {
		return fmt.Errorf("generating commit hook file: %w", err)
	}

	return nil
}

const (
	datasetCreateScript = `#!/bin/sh
# Commit hook installed by beaker alpha commit-hooks dataset-from-src

hash=$(git rev-parse --short HEAD)
pushd {{ .TopDir }} 2>&1 > /dev/null
beaker dataset create {{ if .Workspace -}} -w {{ .Workspace }} {{ end -}} \
	-n {{ .ProjectName | sanitize }}_$hash \
	{{ .Directory }} 
popd 2>&1 > /dev/null
`
)
