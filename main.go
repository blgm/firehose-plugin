package main

import (
	"os"

	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/firehose-plugin/firehose"
)

type NozzlerCmd struct {
	ui terminal.UI
}

func (c *NozzlerCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "FirehosePlugin",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 5,
			Build: 0,
		},
		MinCliVersion: plugin.VersionType{
			Major: 0,
			Minor: 3,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "nozzle",
				HelpText: "Command to print out messages from the firehose",
				UsageDetails: plugin.Usage{
					Usage: "cf nozzle",
					Options: map[string]string{
						"debug": "true to enable debugging",
					},
				},
			},
		},
	}
}

func setupFlags() map[string]flags.FlagSet {
	fs := make(map[string]flags.FlagSet)
	fs["debug"] = &cliFlags.BoolFlag{Name: "debug", Usage: "Used for debugging"}
	fs["no-filter"] = &cliFlags.BoolFlag{Name: "no-filter", Usage: "no firehose filter. Display all messages"}
	fs["filter"] = &cliFlags.StringFlag{Name: "filter", Usage: "Specify message filter such as LogMessage, ValueMetric, CounterEvent, HttpStartStop"}
	fs["subscription-id"] = &cliFlags.StringFlag{Name: "subscription-id", Usage: "Specify subscription id for splitting firehose output between clients"}

	return fs
}

func main() {
	plugin.Start(new(NozzlerCmd))
}

func (c *NozzlerCmd) Run(cliConnection plugin.CliConnection, args []string) {

	if args[0] != "nozzle" {
		return
	}
	c.ui = terminal.NewUI(os.Stdin, terminal.NewTeePrinter())

	dopplerEndpoint, err := cliConnection.DopplerEndpoint()
	if err != nil {
		c.ui.Failed(err.Error())
	}

	authToken, err := cliConnection.AccessToken()
	if err != nil {
		c.ui.Failed(err.Error())
	}

	options := c.buildClientOptions(args)

	client := firehose.NewClient(authToken, dopplerEndpoint, options, c.ui)
	client.Start()
}

func (c *NozzlerCmd) buildClientOptions(args []string) *firehose.ClientOptions {
	var debug bool
	var noFilter bool
	var filter string
	var subscriptionId string

	fc := flags.NewFlagContext(setupFlags())
	err := fc.Parse(args[1:]...)
	if err != nil {
		c.ui.Failed(err.Error())
	}
	if fc.IsSet("debug") {
		debug = fc.Bool("debug")
	}
	if fc.IsSet("no-filter") {
		noFilter = fc.Bool("no-filter")
	}
	if fc.IsSet("filter") {
		filter = fc.String("filter")
	}
	if fc.IsSet("subscription-id") {
		subscriptionId = fc.String("subscription-id")
	}

	return &firehose.ClientOptions{
		Debug:          debug,
		NoFilter:       noFilter,
		Filter:         filter,
		SubscriptionID: subscriptionId,
	}

}
