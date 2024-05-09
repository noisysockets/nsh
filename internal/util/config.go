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

package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/noisysockets/noisysockets/config"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
)

// UpdateConfig performs an atomic update on the given config file.
func UpdateConfig(configPath string, update func(*v1alpha1.Config) (*v1alpha1.Config, error)) error {
	lock := flock.New(configPath + ".lock")
	locked, err := lock.TryLock()
	if err != nil {
		return fmt.Errorf("error acquiring lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("config file is locked by another process")
	}
	defer lock.Unlock()

	configFile, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error opening config file: %w", err)
		}
	} else {
		defer configFile.Close()
	}

	var conf *v1alpha1.Config
	if configFile != nil {
		conf, err = config.FromYAML(configFile)
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
