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
	"github.com/noisysockets/noisysockets/config/v1alpha1"
)

func Export(conf *v1alpha1.Config, wireguardConfigPath string) error {
	var w io.Writer
	if wireguardConfigPath == "-" {
		w = os.Stdout
	} else {
		if err := os.Remove(wireguardConfigPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error removing existing wireguard config file: %w", err)
		}

		wireguardConfigFile, err := os.OpenFile(wireguardConfigPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o400)
		if err != nil {
			return fmt.Errorf("error opening wireguard config: %w", err)
		}
		defer wireguardConfigFile.Close()
		w = wireguardConfigFile
	}

	if err := config.ToINI(w, conf); err != nil {
		return fmt.Errorf("error writing wireguard config: %w", err)
	}

	return nil
}
