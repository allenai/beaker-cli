package config

import (
	"strings"
	"github.com/pkg/errors"
	"reflect"
	"path/filepath"
	"fmt"
	"io/ioutil"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/allenai/beaker/config"
)

type setOptions struct {
	key string
	value string
}

func newSetCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	o := &setOptions{}
	cmd := parent.Command("set", "Set a specific config setting, identified by its YAML key")
	cmd.Action(func(c *kingpin.ParseContext) error {
		return o.run(config)
	})

	cmd.Arg("key", "Key").Required().StringVar(&o.key)
	cmd.Arg("value", "Value").Required().StringVar(&o.value)
}

func (o *setOptions) run(beakerCfg *config.Config) error {
	t := reflect.TypeOf(*beakerCfg)
	found := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("yaml") == o.key {
			found = true
			reflect.ValueOf(beakerCfg).Elem().FieldByName(field.Name).SetString(strings.TrimSpace(o.value))
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Unknown config field: %q", o.key))
	}

	fmt.Printf("Set %s = %s\n", o.key, o.value)

	bytes, err := yaml.Marshal(beakerCfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(config.BeakerConfigDir, os.ModePerm); err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(filepath.Join(config.BeakerConfigDir, "config.yml"), bytes, 0644)
}
