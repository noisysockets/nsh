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

	"github.com/noisysockets/noisysockets/config"
)

func Import(configPath, wireguardConfigPath string) error {
	var r io.Reader
	if wireguardConfigPath == "-" {
		r = os.Stdin
	} else {
		wireguardConfigFile, err := os.Open(wireguardConfigPath)
		if err != nil {
			return fmt.Errorf("error opening wireguard config: %w", err)
		}
		defer wireguardConfigFile.Close()
		r = wireguardConfigFile
	}

	conf, err := config.FromINI(r)
	if err != nil {
		return fmt.Errorf("error parsing wireguard config: %w", err)
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing existing noisy sockets config file: %w", err)
	}

	configFile, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o400)
	if err != nil {
		return fmt.Errorf("error opening noisy sockets config file: %w", err)
	}
	defer configFile.Close()

	if err := config.ToYAML(configFile, conf); err != nil {
		return fmt.Errorf("error writing noisy sockets config: %w", err)
	}

	return nil
}
