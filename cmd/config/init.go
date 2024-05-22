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
	"log/slog"
	"os"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/noisysockets/types"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Init(logger *slog.Logger, configPath string, hostname string,
	listenPort int, ips []string, domain string) error {
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

	return util.UpdateConfig(logger, configPath, func(_ *latestconfig.Config) (*latestconfig.Config, error) {
		conf := &latestconfig.Config{
			Name:       hostname,
			ListenPort: uint16(listenPort),
			PrivateKey: privateKey.String(),
			IPs:        ips,
		}

		if domain != "" {
			conf.DNS = &latestconfig.DNSConfig{
				Domain: domain,
			}
		}

		return conf, nil
	})
}
