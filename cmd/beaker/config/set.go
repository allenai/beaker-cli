package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	yaml "gopkg.in/yaml.v2"

	"github.com/allenai/beaker/config"
)

type setOptions struct {
	property string
	value    string
}

func newSetCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	o := &setOptions{}
	cmd := parent.Command("set", "Set a specific config setting")
	cmd.Action(func(c *kingpin.ParseContext) error {
		return o.run(config)
	})

	cmd.Arg("property", "Name of the property to set").Required().StringVar(&o.property)
	cmd.Arg("value", "New value to set").Required().StringVar(&o.value)
}

func (o *setOptions) run(beakerCfg *config.Config) error {
	t := reflect.TypeOf(*beakerCfg)
	found := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("yaml") == o.property {
			found = true
			// The following code assumes all values are strings and will not work with non-string values.
			reflect.ValueOf(beakerCfg).Elem().FieldByName(field.Name).SetString(strings.TrimSpace(o.value))
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Unknown config property: %q", o.property))
	}

	fmt.Printf("Set %s = %s\n", o.property, o.value)

	bytes, err := yaml.Marshal(beakerCfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(config.BeakerConfigDir, os.ModePerm); err != nil {
		return errors.WithStack(err)
	}

	return ioutil.WriteFile(filepath.Join(config.BeakerConfigDir, "config.yml"), bytes, 0644)
}
