package cli

import (
	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"
)

type Cli struct {
	Version      kong.VersionFlag   `help:"Show version."`
	ConfDir      string             `default:"./containers" help:"Discourse pups config directory." predictor:"dir"`
	TemplatesDir string             `default:"." help:"Home project directory containing a templates/ directory which in turn contains pups yaml templates." predictor:"dir"`
	BuildDir     string             `default:"./tmp" help:"Temporary build folder for building images." predictor:"dir"`
	ForceMkdir   bool               `short:"p" name:"parent-dirs" help:"Create intermediate output directories as required.  If this option is not specified, the full path prefix of each operand must already exist."`
	Upgrade      CliUpgrade         `cmd:"" help:"Upgrade launcher"`
	CliGenerate  CliGenerate        `cmd:"" name:"generate" help:"Generate commands, used to generate Discourse pups, and other Discourse configuration for external tools."`
	BuildCmd     DockerBuildCmd     `cmd:"" name:"build" help:"Build a base image. This command does not need a running database. Saves resulting container."`
	ConfigureCmd DockerConfigureCmd `cmd:"" name:"configure" help:"Configure and save an image with all dependencies and environment baked in. Updates themes and precompiles all assets. Saves resulting container."`
	MigrateCmd   DockerMigrateCmd   `cmd:"" name:"migrate" help:"Run migration tasks for a site. Running container is temporary and is not saved."`
	BootstrapCmd DockerBootstrapCmd `cmd:"" name:"bootstrap" help:"Builds, migrates, and configures an image. Resulting image is a fully built and configured Discourse image."`

	DestroyCmd DestroyCmd `cmd:"" alias:"rm" name:"destroy" help:"Shutdown and destroy container."`
	LogsCmd    LogsCmd    `cmd:"" name:"logs" help:"Print logs for container."`
	CleanupCmd CleanupCmd `cmd:"" name:"cleanup" help:"Cleanup unused containers."`
	EnterCmd   EnterCmd   `cmd:"" name:"enter" help:"Connects to a shell running in the container."`
	RunCmd     RunCmd     `cmd:"" name:"run" help:"Runs the specified command in context of a docker container."`
	StartCmd   StartCmd   `cmd:"" name:"start" help:"Starts container."`
	StopCmd    StopCmd    `cmd:"" name:"stop" help:"Stops container."`
	RestartCmd RestartCmd `cmd:"" name:"restart" help:"Stops then starts container."`
	RebuildCmd RebuildCmd `cmd:"" name:"rebuild" help:"Builds new image, then destroys old container, and starts new container."`

	InstallCompletions kongplete.InstallCompletions `cmd:"" aliases:"sh" help:"Print shell autocompletions. Add output to dotfiles, or 'source <(./launcher2 sh)'."`
	kong.Plugins
}
