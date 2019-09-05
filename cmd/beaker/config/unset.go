package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/allenai/beaker/config"
)

type unsetOptions struct {
	property string
}

func newUnsetCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	o := &unsetOptions{}
	cmd := parent.Command("unset", "Unset a specific config setting")
	cmd.Action(func(c *kingpin.ParseContext) error {
		return o.run(config)
	})

	cmd.Arg("property", "Name of the property to set").Required().StringVar(&o.property)
}

func (o *unsetOptions) run(beakerCfg *config.Config) error {
	t := reflect.TypeOf(*beakerCfg)
	found := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("yaml") == o.property {
			found = true
			reflect.ValueOf(beakerCfg).Elem().FieldByName(field.Name).Set(reflect.Zero(field.Type))
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Unknown config property: %q", o.property))
	}

	fmt.Printf("Unset %s\n", o.property)

	bytes, err := yaml.Marshal(beakerCfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(config.BeakerConfigDir, os.ModePerm); err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(filepath.Join(config.BeakerConfigDir, "config.yml"), bytes, 0644)
}
