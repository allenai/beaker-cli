package config

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

func newListCmd(
	parent *kingpin.CmdClause,
	parentOpts *configOptions,
	config *config.Config,
) {
	cmd := parent.Command("list", "List all configuration properties")
	cmd.Action(func(c *kingpin.ParseContext) error {
		t := reflect.TypeOf(*config)
		fmt.Println("Property\tValue")
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			propertyKey := field.Tag.Get("yaml")
			value := reflect.ValueOf(config).Elem().FieldByName(field.Name).String()
			if value == "" {
				value = "(unset)"
			}
			fmt.Printf("%s\t%s\n", propertyKey, color.BlueString(value))
		}
		return nil
	})
}
