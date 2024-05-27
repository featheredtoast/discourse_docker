package plugin_loader

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/discourse/discourse_docker/launcher_go/v2/cli"
	"os"
	"plugin"
)

type CliPlugin interface {
	Load(*cli.Cli)
}

func LoadPlugins(cli *cli.Cli) {
	cli.Plugins = kong.Plugins{}

	mod := "./launcher_go/v2/sample_plugin/concourse.so"

	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(mod)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 2. look up a symbol (an exported function or variable)
	symPlugin, err := plug.Lookup("CliPlugin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 3. Assert that loaded symbol is of a desired type
	// in this case interface type Greeter (defined above)
	var cliPlugin CliPlugin
	cliPlugin, ok := symPlugin.(CliPlugin)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}

	// 4. use the module
	cliPlugin.Load(cli)
}
