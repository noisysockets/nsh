// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/noisysockets/noisysockets/config"
	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
)

// UpdateConfig performs an atomic update on the given config file.
func UpdateConfig(logger *slog.Logger, configPath string, update func(*latestconfig.Config) (*latestconfig.Config, error)) error {
	lockPath := configPath + ".lock"
	lock := flock.New(lockPath)
	locked, err := lock.TryLock()
	if err != nil {
		return fmt.Errorf("error acquiring lock: %w", err)
	}
	if !locked {
		return errors.New("config file is locked by another process")
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			logger.Error("Error releasing lock", slog.Any("error", err))
		}

		if err := os.Remove(lockPath); err != nil && !os.IsNotExist(err) {
			logger.Error("Error removing lock file", slog.Any("error", err))
		}
	}()

	configFile, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error opening config file: %w", err)
		}
	}

	var conf *latestconfig.Config
	if configFile != nil {
		conf, err = config.FromYAML(configFile)
		_ = configFile.Close()
		if err != nil {
			return fmt.Errorf("error parsing config: %w", err)
		}
	}

	updatedConf, err := update(conf)
	if err != nil {
		return err
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing existing config file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile, err = os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0o400)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}
	defer configFile.Close()

	if err := config.ToYAML(configFile, updatedConf); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	return nil
}
