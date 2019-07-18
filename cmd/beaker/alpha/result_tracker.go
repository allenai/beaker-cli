package alpha

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/beaker/client/api"
)

type datasetPath struct {
	DatasetID string
	Path      string
}

// resultTracker follows a group of experiments in real time and populates paths
// in which to store results for all associated tasks. It's designed to follow
// experiments being added/removed from a group and experiment renames.
type resultTracker struct {
	basePath    string
	experiments map[string]*api.Experiment // Cached experiments keyed by ID
}

func newResultTracker(basePath string) *resultTracker {
	return &resultTracker{basePath, map[string]*api.Experiment{}}
}

// SetExperiments reifies a tracker's cached experiements, adding or removing
// as necessary. It accounts for duplicate experiments and tasks, and will
// remove experiments that do not appear in the given list.
func (t *resultTracker) SetExperiments(experiments []*api.Experiment) error {
	seen := map[string]bool{}
	for _, exp := range experiments {
		if seen[exp.ID] {
			continue
		}
		seen[exp.ID] = true

		cached, ok := t.experiments[exp.ID]
		if !ok {
			if err := t.addExperiment(exp); err != nil {
				return err
			}
			continue
		}

		// No changes necessary if the experiment doesn't differ from the cache.
		if cached.Name == exp.Name {
			continue
		}

		// TODO: It's possible for this to collide when experiments are renamed
		// and one of them takes over an abandoned name. We should guarantee a
		// safe order or retry.
		if err := t.moveExperiment(cached, exp); err != nil {
			return err
		}
		t.experiments[exp.ID] = exp
	}

	// Drop any deleted/removed experiments.
	for _, exp := range t.experiments {
		if seen[exp.ID] {
			continue
		}

		if err := t.removeExperiment(exp); err != nil {
			return err
		}
	}
	return nil
}

func (t *resultTracker) addExperiment(exp *api.Experiment) error {
	experimentPath := filepath.Join(t.basePath, exp.DisplayID())
	for _, node := range exp.Nodes {
		nodePath := filepath.Join(experimentPath, node.DisplayID())
		if err := os.MkdirAll(nodePath, 0755); err != nil {
			return errors.WithStack(err)
		}
	}
	t.experiments[exp.ID] = exp
	return nil
}

func (t *resultTracker) removeExperiment(exp *api.Experiment) error {
	if err := os.RemoveAll(filepath.Join(t.basePath, exp.DisplayID())); err != nil {
		return err
	}
	delete(t.experiments, exp.ID)
	return nil
}

func (t *resultTracker) moveExperiment(oldExp *api.Experiment, newExp *api.Experiment) error {
	oldPath := filepath.Join(t.basePath, oldExp.DisplayID())
	newPath := filepath.Join(t.basePath, newExp.DisplayID())
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	t.experiments[newExp.ID] = newExp
	return nil
}

func (t *resultTracker) DatasetPaths() []datasetPath {
	var result []datasetPath
	for _, exp := range t.experiments {
		for _, node := range exp.Nodes {
			result = append(result, datasetPath{
				DatasetID: node.ResultID,
				Path:      filepath.Join(t.basePath, exp.DisplayID(), node.DisplayID()),
			})
		}
	}
	return result
}
