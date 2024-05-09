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
	"os"
	"path/filepath"

	"github.com/noisysockets/noisysockets/config"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
)

func Export(conf *v1alpha1.Config, wireGuardConfigPath string) error {
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
