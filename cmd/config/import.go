// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/noisysockets/noisysockets/config"
	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/nsh/internal/util"
)

func Import(logger *slog.Logger, configPath, wireGuardConfigPath string) error {
	var r io.Reader
	if wireGuardConfigPath == "-" {
		r = os.Stdin
	} else {
		wireGuardConfigFile, err := os.Open(wireGuardConfigPath)
		if err != nil {
			return fmt.Errorf("error opening WireGuard config: %w", err)
		}
		defer wireGuardConfigFile.Close()
		r = wireGuardConfigFile
	}

	return util.UpdateConfig(logger, configPath, func(_ *latestconfig.Config) (*latestconfig.Config, error) {
		conf, err := config.FromINI(r)
		if err != nil {
			return nil, fmt.Errorf("error parsing WireGuard config: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		return conf, nil
	})
}
