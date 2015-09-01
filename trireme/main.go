package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/deis/deis/trireme/platform"
	"github.com/deis/deis/trireme/storage"
)

// Components, in the order in which they must be started.
//
// This list describes all of the components that the CLI has the ability to
// manage. When things are installed, namespaces, services, volumes and secrets
// are installed first, then pods and rcs are installed.
var components = []*platform.Component{
	{
		Name:        "builder",
		Description: "the builder",
		Namespaces:  []string{"namespaces/deis-namespace.json"},
		RCs:         []string{"rcs/deis-builder.json"},
		Services:    []string{"services/deis-builder-service.json"},
	},
	{
		Name:        "controller",
		Description: "the controller",
		RCs:         []string{"rcs/deis-controller.json"},
		Services:    []string{"services/deis-controller-service.json"},
	},
	{
		Name:        "database",
		Description: "the database",
		RCs:         []string{"rcs/deis-database.json"},
		Services:    []string{"services/deis-database-service.json"},
	},
	{
		Name:        "registry",
		Description: "the registry",
		RCs:         []string{"rcs/deis-registry.json"},
		Services:    []string{"services/deis-registry-service.json"},
	},
	{
		Name:        "router",
		Description: "the router mesh",
		RCs:         []string{"rcs/deis-router.json"},
		//Services:    []string{"services/deis-router-service.json"},
	},
	{
		Name:        "store",
		Description: "the persistent storage cluster",
		RCs:         []string{"rcs/deis-store-gtw.json", "rcs/deis-store-mds.json", "rcs/deis-store-mon.json", "rcs/deis-store-osd.json"},
		Services:    []string{"services/deis-store-gtw-service.json"},
		Optional:    true,
	},
}

// config is where configuration data is stored.
var config storage.Storer

func main() {

	// This is temporary: We need a place to store configuration data in
	// lieu of etcd. For now, we'll store it locally in a configuration
	// file.
	var err error
	config, err = storage.New(defaultConfigFile())
	if err != nil {
		fmt.Printf("Failed to load or create file %s", defaultConfigFile())
		os.Exit(321)
	}

	app := cli.NewApp()
	app.Name = "trireme"
	app.Usage = "A control tool for Deis"
	app.Flags = []cli.Flag{
		// This is currently not used.
		cli.StringFlag{
			Name:   "tunnel, t",
			Value:  "127.0.0.1",
			EnvVar: "DEISCTL_TUNNEL",
			Usage:  "The IP address or hostname of the tunnel.",
		},
	}
	app.Commands = commands()
	app.Run(os.Args)
}

func commands() []cli.Command {
	return []cli.Command{
		{
			Name:        "config",
			Usage:       "Get and set configuration values",
			Subcommands: configCommands(),
		},
		{
			Name:        "install",
			Usage:       "Install platform components",
			Subcommands: installCommands(),
		},
		{
			Name:        "uninstall",
			Usage:       "Uninstall platform components",
			Subcommands: uninstallCommands(),
		},
	}
}

func defaultConfigFile() string {
	return os.ExpandEnv("${HOME}/.trireme")
}

func configCommands() []cli.Command {
	cmds := make([]cli.Command, 0, len(components)+3)

	// Why deprecate `deisctl config <target> set ...`? Three reasons:
	// 1. The predominant form of multi-commands is <CMD> <VERB> <DO>..., not
	//    <CMD> <NOUN> <DO> <VERB>...
	// 2. The common unix paradigm for subcommands is to move all variables to
	//    the arguments portion: <CMD> <SUBCMD> <ARG1>..., not <CMD> <SUBCMD> <ARG1> <SUBCMD> <ARG2>...
	// 3. The logic is simply cleaner when arguments are grouped together.
	cmds = append(cmds,
		cli.Command{
			Name:  "get",
			Usage: "Get an existing parameter from an existing component",
			Action: func(c *cli.Context) {
				a := c.Args()
				if len(a) < 2 {
					fmt.Println("Usage: deisctl config get <TARGET> <KEY>")
					os.Exit(1)
				}
				ns := a[0]
				key := a[1]
				val, err := config.Get(ns, key)
				if err != nil {
					fmt.Printf("Failed to get %s:%s: '%s'\n", ns, key, err)
					os.Exit(2)
				}
				fmt.Println(val)
			},
		},
		cli.Command{
			Name:  "set",
			Usage: "Set an existing parameter for an existing component",
			Action: func(c *cli.Context) {
				a := c.Args()
				if len(a) < 3 {
					fmt.Println("Usage: deisctl config get <TARGET> <KEY> <VALUE>")
					os.Exit(1)
				}
				ns := a[0]
				key := a[1]
				val := a[2]
				if err := config.Set(ns, key, val); err != nil {
					fmt.Printf("Failed to set %s:%s=%s: '%s'\n", ns, key, val, err)
					os.Exit(3)
				}
				fmt.Println(val)
			},
		},
		cli.Command{
			Name:  "rm",
			Usage: "Remove an existing parameter from an existing component",
			Action: func(c *cli.Context) {
				a := c.Args()
				if len(a) < 2 {
					fmt.Println("Usage: deisctl config rm <TARGET> <KEY>")
					os.Exit(1)
				}
				ns := a[0]
				key := a[1]
				if err := config.Remove(ns, key); err != nil {
					fmt.Printf("Failed to remove %s:%s: '%s'\n", ns, key, err)
					os.Exit(4)
				}
			},
		},
	)

	for _, c := range components {
		n := c.Name
		cmds = append(cmds, cli.Command{
			Name:  n,
			Usage: fmt.Sprintf("Deprecated. Use `deisctl config set|get|rm %s`", n),
			Action: func(c *cli.Context) {
				a := c.Args()
				if len(a) < 2 {
					fmt.Println("Usage: deisctl config get|set|rm <TARGET> <KEY> [<VALUE>]")
					os.Exit(1)
				}
				key := a[1]
				switch a[0] {
				case "get":
					val, err := config.Get(n, key)
					if err != nil {
						fmt.Println(err)
						os.Exit(2)
					}
					fmt.Println(val)
				case "set":
					if err := config.Set(n, key, a[2]); err != nil {
						fmt.Println("Usage: deisctl config get|set|rm <TARGET> <KEY> [<VALUE>]")
						os.Exit(3)
					}
					fmt.Println(a[2])
				case "rm":
					if err := config.Remove(n, key); err != nil {
						fmt.Println("Usage: deisctl config get|set|rm <TARGET> <KEY> [<VALUE>]")
						os.Exit(4)
					}
				default:
					fmt.Println("Usage: deisctl config get|set|rm <TARGET> <KEY> [<VALUE>]")
					os.Exit(1)
				}
			},
		})
	}
	return cmds
}

func installCommands() []cli.Command {
	// This basically ensures that append() will not have to reallocate.
	cmds := make([]cli.Command, 0, len(components)+1)

	for _, c := range components {
		n := c.Name
		v := c.Description
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
				Name:   "unit-files, u",
				Value:  "./units/",
				Usage:  "The path to the Deis Kubernetes JSON unit files.",
				EnvVar: "DEISCTL_UNITS",
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
func uninstallCommands() []cli.Command {
	// This basically ensures that append() will not have to reallocate.
	cmds := make([]cli.Command, 0, len(components)+1)

	for _, c := range components {
		n := c.Name
		v := c.Description
		cmds = append(cmds, cli.Command{
			Name:   n,
			Usage:  fmt.Sprintf("Uninstall %s.", v),
			Action: func(c *cli.Context) { uninstallComponent(c, n) },
		})
	}
	cmds = append(cmds, cli.Command{
		Name:   "platform",
		Usage:  "Uninstall the entire platform",
		Action: uninstallPlatform,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "unit-files, u",
				Value:  "./units/",
				Usage:  "The path to the Deis Kubernetes JSON unit files.",
				EnvVar: "DEISCTL_UNITS",
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

	dir := c.String("unit-files")
	if err := platform.InstallAll(components, dir, false); err != nil {
		fmt.Printf("Failed to install platform: %s\n", err)
		os.Exit(500)
	}
}

func uninstallPlatform(c *cli.Context) {
	dir := c.String("unit-files")
	if err := platform.DeleteAll(components, dir); err != nil {
		fmt.Printf("Failed to uninstall platform: %s\n", err)
		os.Exit(500)
	}
}
func installComponent(c *cli.Context, name string) {
	comp, err := findComponent(name)
	if err != nil {
		fmt.Printf("Failed to install %s: %s\n", name, err)
		os.Exit(404)
	}
	dir := c.String("unit-files")
	if err := comp.Install(dir); err != nil {
		fmt.Printf("Failed to install %s: %s\n", name, err)
		os.Exit(500)
	}
	if err := comp.InstallPrereqs(dir); err != nil {
		fmt.Printf("Failed to install dependencies of %s: %s\n", name, err)
		os.Exit(500)
	}
}

func uninstallComponent(c *cli.Context, name string) {
	comp, err := findComponent(name)
	if err != nil {
		fmt.Printf("Failed to uninstall %s: %s\n", name, err)
		os.Exit(404)
	}
	dir := c.String("unit-files")
	if err := comp.Delete(dir); err != nil {
		fmt.Printf("Failed to uninstall %s: %s\n", name, err)
		os.Exit(500)
	}
	if err := comp.DeletePrereqs(dir); err != nil {
		fmt.Printf("Failed to uninstall dependencies of %s: %s\n", name, err)
		os.Exit(500)
	}
}

func findComponent(name string) (*platform.Component, error) {
	for _, c := range components {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, errors.New("no component found")
}
