// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"runtime"
	"time"

	"github.com/adrg/xdg"
	"github.com/dpeckett/telemetry"
	"github.com/dpeckett/telemetry/v1alpha1"
	"github.com/noisysockets/network"
	"github.com/noisysockets/noisysockets/config"
	configtypes "github.com/noisysockets/noisysockets/config/types"
	configcmd "github.com/noisysockets/nsh/cmd/config"
	dnscmd "github.com/noisysockets/nsh/cmd/dns"
	peercmd "github.com/noisysockets/nsh/cmd/peer"
	routecmd "github.com/noisysockets/nsh/cmd/route"
	upcmd "github.com/noisysockets/nsh/cmd/up"
	"github.com/noisysockets/nsh/internal/constants"
	"github.com/noisysockets/nsh/internal/service"
	"github.com/noisysockets/nsh/internal/util"

	"github.com/urfave/cli/v2"
)

func main() {
	var conf configtypes.Config
	configPath, err := xdg.ConfigFile("nsh/noisysockets.yaml")
	if err != nil {
		slog.Error("Error getting config file path", slog.Any("error", err))
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
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*util.LevelFlag)),
		})))

		return nil
	}

	loadConfig := func(c *cli.Context) error {
		configPath := c.String("config")

		slog.Debug("Loading config", slog.String("path", configPath))

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

	// Collect anonymous usage statistics.
	var telemetryReporter *telemetry.Reporter

	initTelemetry := func(c *cli.Context) error {
		// Compatibility with the old environment variable.
		if os.Getenv("NSH_NO_TELEMETRY") != "" {
			os.Setenv("DO_NOT_TRACK", "1")
		}

		telemetryReporter = telemetry.NewReporter(c.Context, slog.Default(), telemetry.Configuration{
			BaseURL: constants.TelemetryURL,
			Tags:    []string{"nsh"},
		})

		// Some basic system information.
		info := map[string]string{
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"num_cpu": fmt.Sprintf("%d", runtime.NumCPU()),
			"version": constants.Version,
		}

		// Are we running in a container?
		if os.Getenv("container") != "" {
			info["container"] = os.Getenv("container")
		}

		telemetryReporter.ReportEvent(&v1alpha1.TelemetryEvent{
			Kind:   v1alpha1.TelemetryEventKindInfo,
			Name:   "ApplicationStart",
			Values: info,
		})

		return nil
	}

	shutdownTelemetry := func(c *cli.Context) error {
		if telemetryReporter == nil {
			return nil
		}

		telemetryReporter.ReportEvent(&v1alpha1.TelemetryEvent{
			Kind: v1alpha1.TelemetryEventKindInfo,
			Name: "ApplicationStop",
		})

		// Don't want to block the shutdown of the application for too long.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := telemetryReporter.Shutdown(ctx); err != nil {
			slog.Error("Failed to close telemetry reporter", slog.Any("error", err))
		}

		return nil
	}

	app := &cli.App{
		Name:    "nsh",
		Usage:   "The Noisy Sockets CLI",
		Version: constants.Version,
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
							},
							&cli.StringSliceFlag{
								Name:  "ip",
								Usage: "The IP address/s to assign to the peer, if not set a random IPv6 address will be assigned",
							},
							&cli.StringFlag{
								Name:    "domain",
								Aliases: []string{"d"},
								Usage:   "The network domain",
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, initTelemetry),
						After:  shutdownTelemetry,
						Action: func(c *cli.Context) error {
							return configcmd.Init(
								c.String("config"),
								c.String("name"),
								c.Int("listen-port"),
								c.StringSlice("ip"),
								c.String("domain"))
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
						Before: beforeAll(initLogger, initTelemetry),
						After:  shutdownTelemetry,
						Action: func(c *cli.Context) error {
							return configcmd.Import(
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
							&cli.BoolFlag{
								Name:  "stripped",
								Usage: "Remove wg-quick specific fields",
								Value: false,
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, initTelemetry, loadConfig),
						After:  shutdownTelemetry,
						Action: func(c *cli.Context) error {
							return configcmd.Export(
								conf,
								c.String("output"),
								c.Bool("stripped"))
						},
					},
					{
						Name:      "show",
						Usage:     "Show the current configuration",
						Flags:     sharedFlags,
						Args:      true,
						ArgsUsage: "query",
						Before:    beforeAll(initLogger, initTelemetry, loadConfig),
						After:     shutdownTelemetry,
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return errors.New("expected jq syntax query as argument")
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
						Before: beforeAll(initLogger, initTelemetry, loadConfig),
						After:  shutdownTelemetry,
						Action: func(c *cli.Context) error {
							return peercmd.Add(
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
						Before:    beforeAll(initLogger, initTelemetry, loadConfig),
						After:     shutdownTelemetry,
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return errors.New("expected name or public-key as argument")
							}

							return peercmd.Remove(
								c.String("config"),
								c.Args().First(),
							)
						},
					},
				},
			},
			{
				Name:  "route",
				Usage: "Manage network routes",
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
								Usage:    "The router peer name or public key",
								Required: true,
							},
						}, sharedFlags...),
						Before: beforeAll(initLogger, initTelemetry, loadConfig),
						After:  shutdownTelemetry,
						Action: func(c *cli.Context) error {
							return routecmd.Add(
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
						Before:    beforeAll(initLogger, initTelemetry, loadConfig),
						After:     shutdownTelemetry,
						Action: func(c *cli.Context) error {
							if c.Args().Len() != 1 {
								_ = cli.ShowSubcommandHelp(c)
								return errors.New("expected destination as argument")
							}

							return routecmd.Remove(
								c.String("config"),
								c.Args().First(),
							)
						},
					},
				},
			},
			{
				Name:  "dns",
				Usage: "Manage DNS configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "server",
						Usage: "Manage DNS servers",
						Subcommands: []*cli.Command{
							{
								Name:      "add",
								Usage:     "Add a DNS server",
								Args:      true,
								ArgsUsage: "address",
								Flags:     sharedFlags,
								Before:    beforeAll(initLogger, initTelemetry, loadConfig),
								After:     shutdownTelemetry,
								Action: func(c *cli.Context) error {
									if c.Args().Len() != 1 {
										_ = cli.ShowSubcommandHelp(c)
										return errors.New("expected DNS server address as argument")
									}

									return dnscmd.AddServer(
										c.String("config"),
										c.Args().First(),
									)
								},
							},
						},
					},
				},
			},
			{
				Name:  "up",
				Usage: "Start Noisy Sockets",
				Flags: append([]cli.Flag{
					&cli.BoolFlag{
						Name:  "enable-dns",
						Usage: "Enable DNS service",
					},
					&cli.BoolFlag{
						Name:  "enable-router",
						Usage: "Enable router service",
					},
					&cli.BoolFlag{
						Name:  "nat64",
						Usage: "Enable DNS64/NAT64 (IPv6 to IPv4 translation)",
						Value: true,
					},
					&cli.StringFlag{
						Name:  "nat64-prefix",
						Usage: "The DNS64/NAT64 prefix",
						Value: "64:ff9b::/96",
					},
					&cli.StringSliceFlag{
						Name:  "dns-public-upstream",
						Usage: "Upstream DNS servers to use for public queries",
					},
				}, sharedFlags...),
				Before: beforeAll(initLogger, initTelemetry, loadConfig),
				After:  shutdownTelemetry,
				Action: func(c *cli.Context) error {
					enableNAT64 := c.Bool("nat64")

					nat64Prefix, err := netip.ParsePrefix(c.String("nat64-prefix"))
					if err != nil {
						return fmt.Errorf("failed to parse NAT64 prefix: %w", err)
					}

					var services []service.Service

					if c.Bool("enable-dns") {
						services = append(services, service.DNS(enableNAT64, nat64Prefix, c.StringSlice("dns-public-upstream")))
					}

					if c.Bool("enable-router") {
						services = append(services, service.Router(network.Host(), enableNAT64, nat64Prefix))
					}

					// If all services are disabled, then throw an error.
					if len(services) == 0 {
						_ = cli.ShowSubcommandHelp(c)
						return errors.New("at least one service must be enabled")
					}

					return upcmd.Up(c.Context, conf, services)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("Error", slog.Any("error", err))
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
