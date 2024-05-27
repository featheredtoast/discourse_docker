package main

import (
	"context"
	"fmt"
	"github.com/alecthomas/kong"
	pluginLoader "github.com/discourse/discourse_docker/launcher_go/v2/plugin_loader"
	"github.com/discourse/discourse_docker/launcher_go/v2/utils"
	"github.com/discourse/discourse_docker/launcher_go/v2/cli"
	"github.com/posener/complete"
	"github.com/willabides/kongplete"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
)

func main() {
	cli := cli.Cli{}

	pluginLoader.LoadPlugins(&cli)

	runCtx, cancel := context.WithCancel(context.Background())

	// pre parse to get config dir for prediction of conf dir
	confFiles := utils.FindConfigNames()

	parser := kong.Must(&cli, kong.UsageOnError(), kong.Bind(&runCtx), kong.Vars{"version": utils.Version})

	// Run kongplete.Complete to handle completion requests
	kongplete.Complete(parser,
		kongplete.WithPredictor("config", complete.PredictSet(confFiles...)),
		kongplete.WithPredictor("file", complete.PredictFiles("*")),
		kongplete.WithPredictor("dir", complete.PredictDirs("*")),
	)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, unix.SIGTERM)
	signal.Notify(sigChan, unix.SIGINT)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-sigChan:
			fmt.Fprintln(utils.Out, "Command interrupted")
			cancel()
		case <-done:
		}
	}()
	err = ctx.Run()
	if err == nil {
		return
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		// Magic exit code that indicates a retry
		if exiterr.ExitCode() == 77 {
			os.Exit(77)
		} else if runCtx.Err() != nil {
			fmt.Fprintln(utils.Out, "Aborted with exit code", exiterr.ExitCode())
		} else {
			ctx.Fatalf(
				"run failed with exit code %v\n"+
					"** FAILED TO BOOTSTRAP ** please scroll up and look for earlier error messages, there may be more than one.\n"+
					"./discourse-doctor may help diagnose the problem.", exiterr.ExitCode())
		}
	} else {
		ctx.FatalIfErrorf(err)
	}
}
