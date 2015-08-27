package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "trireme"
	app.Usage = "A control tool for Deis"
	app.Commands = commands()
	app.Run(os.Args)
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:  "install",
			Usage: "Install platform components",
			/*
				Action: func(c *cli.Context) {
					fmt.Println("I can only install Platform right now.")
				},
			*/
			Subcommands: installCommands(),
		},
	}
}

func installCommands() []cli.Command {
	parts := map[string]string{
		// In alpha order, please.
		"builder":    "the builder",
		"controller": "the controller",
		"database":   "the database",
		"router":     "the router mesh",
		"store":      "the store and all of its components",
		// FIXME: Do the rest.
	}

	// This basically ensures that append() will not have to reallocate.
	cmds := make([]cli.Command, 0, len(parts)+1)

	for n, v := range parts {
		cmds = append(cmds, cli.Command{
			Name:   n,
			Usage:  fmt.Sprintf("Install %s.", v),
			Action: func(c *cli.Context) { installComponent(c, n) },
		})
	}
	cmds = append(cmds, cli.Command{
		Name:   "platform",
		Usage:  "Install the entire platform",
		Action: installPlatform,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "units, u",
				Value:  "units/",
				Usage:  "The path to the Deis Kubernetes JSON unit files.",
				EnvVar: "DEIS_K8S_UNITS",
			},
			cli.StringFlag{
				Name:   "registry, r",
				Usage:  "The URL to a Docker registry that holds Deis images images.",
				EnvVar: "DEV_REGISTRY",
			},
		},
	})

	return cmds
}

func installPlatform(c *cli.Context) {
	fmt.Println("Installing platform")
}
func installComponent(c *cli.Context, name string) {
	fmt.Printf("Installing component '%s' is currently not supported.\n", name)
	os.Exit(1)
}
