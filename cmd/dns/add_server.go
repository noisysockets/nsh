// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package dns

import (
	"errors"
	"fmt"
	"log/slog"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func AddServer(logger *slog.Logger, configPath, address string) error {
	return util.UpdateConfig(logger, configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		if conf.DNS == nil {
			conf.DNS = &latestconfig.DNSConfig{}
		}

		// Do we already have a server with this address?
		for _, existingAddr := range conf.DNS.Servers {
			if existingAddr == address {
				return nil, errors.New("server already exists")
			}
		}

		if err := validate.IP(address); err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}

		// Add the new server.
		conf.DNS.Servers = append(conf.DNS.Servers, address)

		return conf, nil
	})
}
