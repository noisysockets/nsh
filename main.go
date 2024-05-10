/*
 * Copyright 2024 Damian Peckett <damian@pecke.tt>
 *
 * Licensed under the Noisy Sockets Source License 1.0 (NSSL-1.0); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at
 *
 * https://github.com/noisysockets/nsh/blob/main/LICENSE
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/adrg/xdg"
	"github.com/noisysockets/noisysockets/config"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
	configcmd "github.com/noisysockets/nsh/cmd/config"
	peercmd "github.com/noisysockets/nsh/cmd/peer"
	routecmd "github.com/noisysockets/nsh/cmd/route"
	shellcmd "github.com/noisysockets/nsh/cmd/shell"
	"github.com/noisysockets/nsh/internal/util"

	"github.com/urfave/cli/v2"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var conf *v1alpha1.Config
	configPath, err := xdg.ConfigFile("nsh/noisysockets.yaml")
	if err != nil {
		logger.Error("Error getting config file path", slog.Any("error", err))
		os.Exit(1)
	}

	sharedFlags := []cli.Flag{
		&cli.GenericFlag{
			Name:  "log-level",
			Usage: "Set the log verbosity level",
			Value: util.FromSlogLevel(slog.LevelInfo),
		},
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Noisy Sockets configuration file",
			Value:   configPath,
		},
	}

	initLogger := func(c *cli.Context) error {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*util.LevelFlag)),
		}))

		return nil
	}

	loadConfig := func(c *cli.Context) error {
		configPath := c.String("config")

		logger.Debug("Loading config", slog.String("path", configPath))

		configFile, err := os.Open(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("config file %q does not exist, run `nsh config init` to create one", configPath)
			}

			return fmt.Errorf("failed to open config file: %w", err)
		}
		defer configFile.Close()

		conf, err = config.FromYAML(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		return nil
	}

	app := &cli.App{
		Name:  "nsh",
		Usage: "The Noisy Sockets CLI",
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Manage configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "Create a new configuration",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "The name of the peer",
							},
							&cli.IntFlag{
								Name:    "listen-port",
								Aliases: []string{"l"},
								Usage:   "The port to listen on",
								Value:   51820,
							},
							&cli.StringSliceFlag{
								Name:  "ip",
								Usage: "The IP address/s to assign to the peer",
								// Use the	172.21.248.0/24 subnet as the default.
								// This CIDR is chosen to reduce the likelihood of conflicts.
								Value: cli.NewStringSlice("172.21.248.1"),
							},
						}, sharedFlags...),
						Before: initLogger,
						Action: func(c *cli.Context) error {
							return configcmd.Init(logger,
								c.String("config"),
								c.String("name"),
								c.Int("listen-port"),
								c.StringSlice("ip"))
						},
					},
					{
						Name:  "import",
						Usage: "Import existing WireGuard configuration",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "input",
								Aliases: []string{"i"},
								Usage:   "The path to read the WireGuard formatted configuration",
								Value:   "-",
							},
						}, sharedFlags...),
						Before: initLogger,
						Action: func(c *cli.Context) error {
							return configcmd.Import(
								logger,
								c.String("config"),
								c.String("input"))
						},
					},
					{
						Name:  "export",
						Usage: "Export WireGuard configuration",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
								Usage:   "The path to write the WireGuard formatted configuration",
								Value:   "-",
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							return configcmd.Export(
								conf,
								c.String("output"))
						},
					},
					{
						Name:      "show",
						Usage:     "Show the current configuration",
						Flags:     sharedFlags,
						Args:      true,
						ArgsUsage: "query",
						Before:    beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return fmt.Errorf("expected jq syntax query as argument")
							}

							return configcmd.Show(c.Context, conf, c.Args().First())
						},
					},
				},
			},
			{
				Name:  "peer",
				Usage: "Manage peers",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Add a peer",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								Aliases: []string{"n"},
								Usage:   "The name of the peer",
							},
							&cli.StringFlag{
								Name:     "public-key",
								Aliases:  []string{"k"},
								Usage:    "The public key of the peer",
								Required: true,
							},
							&cli.StringFlag{
								Name:    "endpoint",
								Aliases: []string{"e"},
								Usage:   "The peer's public address/port (if available)",
							},
							&cli.StringSliceFlag{
								Name:     "ip",
								Usage:    "The IP address/s to assign to the peer",
								Required: true,
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							return peercmd.Add(
								logger,
								c.String("config"),
								c.String("name"),
								c.String("public-key"),
								c.String("endpoint"),
								c.StringSlice("ip"),
							)
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a peer",
						Flags:     sharedFlags,
						Args:      true,
						ArgsUsage: "name | public-key",
						Before:    beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return fmt.Errorf("expected name or public-key as argument")
							}

							return peercmd.Remove(
								logger,
								c.String("config"),
								c.Args().First(),
							)
						},
					},
				},
			},
			{
				Name:  "route",
				Usage: "Manage routes",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "Add a route",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:     "destination",
								Aliases:  []string{"d"},
								Usage:    "The destination CIDR",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "via",
								Aliases:  []string{"v"},
								Usage:    "The gateway peer name or public key",
								Required: true,
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							return routecmd.Add(
								logger,
								c.String("config"),
								c.String("destination"),
								c.String("via"),
							)
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a route",
						Flags:     sharedFlags,
						Args:      true,
						ArgsUsage: "destination",
						Before:    beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return fmt.Errorf("expected destination as argument")
							}

							return routecmd.Remove(
								logger,
								c.String("config"),
								c.Args().First(),
							)
						},
					},
				},
			},
			{
				Name:  "shell",
				Usage: "Remote shell",
				Subcommands: []*cli.Command{
					{
						Name:      "connect",
						Usage:     "Connect to a remote shell server",
						Flags:     sharedFlags,
						Args:      true,
						ArgsUsage: "address",
						Before:    beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return fmt.Errorf("expected server address/port as argument")
							}

							return shellcmd.Connect(c.Context, logger, conf, c.Args().First())
						},
					},
					{
						Name:   "serve",
						Usage:  "Start a remote shell server",
						Flags:  sharedFlags,
						Before: beforeAll(initLogger, loadConfig),
						Action: func(c *cli.Context) error {
							return shellcmd.Serve(c.Context, logger, conf)
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Error", slog.Any("error", err))

		var e util.ExitError
		if errors.As(err, &e) {
			os.Exit(int(e))
		}
		os.Exit(1)
	}
}

func beforeAll(funcs ...cli.BeforeFunc) cli.BeforeFunc {
	return func(c *cli.Context) error {
		for _, f := range funcs {
			if err := f(c); err != nil {
				return err
			}
		}

		return nil
	}
}
