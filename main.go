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
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/noisysockets/noisysockets/config"
	"github.com/noisysockets/noisysockets/config/types"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
	configcmd "github.com/noisysockets/nsh/cmd/config"
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

	beforeAll := func(c *cli.Context) error {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*util.LevelFlag)),
		}))

		configPath := c.String("config")

		_, err = os.Stat(configPath)
		configDoesNotExist := os.IsNotExist(err)

		if !c.IsSet("config") && configDoesNotExist {
			configDir := filepath.Dir(configPath)
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				if err := os.MkdirAll(configDir, 0o700); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}
			}

			configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, 0o400)
			if err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
			defer configFile.Close()

			// TODO: generate valid default config
			conf = &v1alpha1.Config{
				TypeMeta: types.TypeMeta{
					APIVersion: v1alpha1.ApiVersion,
					Kind:       "Config",
				},
			}

			if err := config.ToYAML(configFile, conf); err != nil {
				return fmt.Errorf("failed to write default config: %w", err)
			}
		} else {
			configFile, err := os.Open(configPath)
			if err != nil {
				return fmt.Errorf("failed to open config file: %w", err)
			}
			defer configFile.Close()

			conf, err = config.FromYAML(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}
		}

		return nil
	}

	app := &cli.App{
		Name:   "nsh",
		Usage:  "The Noisy Sockets CLI",
		Flags:  sharedFlags,
		Before: beforeAll,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Manage WireGuard configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "import",
						Usage: "Import WireGuard INI configuration",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "input",
								Aliases: []string{"i"},
								Usage:   "The path to read the wireguard configuration",
								Value:   "-",
							},
						}, sharedFlags...),
						Before: beforeAll,
						Action: func(c *cli.Context) error {
							configPath := c.String("config")

							wireguardConfigPath := c.String("input")

							return configcmd.Import(configPath, wireguardConfigPath)
						},
					},
					{
						Name:  "export",
						Usage: "Export WireGuard INI configuration",
						Flags: append([]cli.Flag{
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
								Usage:   "The path to write the configuration",
								Value:   "-",
							},
						}, sharedFlags...),
						Before: beforeAll,
						Action: func(c *cli.Context) error {
							wireguardConfigPath := c.String("output")

							return configcmd.Export(conf, wireguardConfigPath)
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Error", slog.Any("error", err))
		os.Exit(1)
	}
}
