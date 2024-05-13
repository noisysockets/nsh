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
	"os"
	"path/filepath"

	"github.com/noisysockets/noisysockets/config"
	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
)

func Export(conf *latestconfig.Config, wireGuardConfigPath string) error {
	var w io.Writer
	if wireGuardConfigPath == "-" {
		w = os.Stdout
	} else {
		if err := os.Remove(wireGuardConfigPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing existing WireGuard config: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(wireGuardConfigPath), 0o700); err != nil {
			return fmt.Errorf("error creating WireGuard config directory: %w", err)
		}

		wireGuardConfigFile, err := os.OpenFile(wireGuardConfigPath, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("error opening WireGuard config: %w", err)
		}
		defer wireGuardConfigFile.Close()

		w = wireGuardConfigFile
	}

	if err := config.ToINI(w, conf); err != nil {
		return fmt.Errorf("error writing WireGuard config: %w", err)
	}

	return nil
}
