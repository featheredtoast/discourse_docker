package main

import (
	"errors"
	"fmt"
	"github.com/discourse/discourse_docker/launcher_go/v2/cli"
	"github.com/discourse/discourse_docker/launcher_go/v2/config"
	"github.com/discourse/discourse_docker/launcher_go/v2/utils"
)

type ConcourseJobCmd struct {
	Output string `help:"write concourse job to output file"`
	Config string `arg:"" name:"config" help:"config" predictor:"config"`
}

func (r *ConcourseJobCmd) Run(cli *cli.Cli) error {
	fmt.Fprintln(utils.Out, "## WARNING: concourse job generation is experimental, use at your own risk!")
	loadedConfig, err := config.LoadConfig(cli.ConfDir, r.Config, true, cli.TemplatesDir)
	if err != nil {
		return errors.New("YAML syntax error. Please check your containers/*.yml config files.")
	}
	if r.Output == "" {
		fmt.Fprint(utils.Out, GenConcourseConfig(*loadedConfig))
	} else {
		WriteConcourseConfig(*loadedConfig, r.Output)
	}
	return nil
}

type ConcoursePluginCli struct {
	Concourse ConcourseJobCmd `cmd:"" help:"print concourse job info"`
}

type PluginLoader struct {}
func(p PluginLoader) Load(cli *cli.Cli) {
	cli.Plugins = append(cli.Plugins, &ConcoursePluginCli{})
}

var CliPlugin PluginLoader
