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

package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/noisysockets/noisysockets/config"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
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

	return util.UpdateConfig(logger, configPath, func(_ *v1alpha1.Config) (*v1alpha1.Config, error) {
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
