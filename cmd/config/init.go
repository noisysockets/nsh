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
	"log/slog"
	"os"

	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/noisysockets/types"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Init(logger *slog.Logger, configPath string, hostname string, listenPort int, ips []string) error {
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			logger.Warn("Error getting hostname", slog.Any("error", err))
		}
	}

	privateKey, err := types.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	if err := validate.IPs(ips); err != nil {
		return fmt.Errorf("invalid IP address: %w", err)
	}

	return util.UpdateConfig(logger, configPath, func(_ *v1alpha1.Config) (*v1alpha1.Config, error) {
		return &v1alpha1.Config{
			Name:       hostname,
			ListenPort: uint16(listenPort),
			PrivateKey: privateKey.String(),
			IPs:        ips,
		}, nil
	})
}
