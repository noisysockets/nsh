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

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha3"
	"github.com/noisysockets/noisysockets/types"
	"github.com/noisysockets/nsh/internal/util"
)

func AddServer(configPath, address string) error {
	return util.UpdateConfig(configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		if conf.DNS == nil {
			conf.DNS = &latestconfig.DNSConfig{}
		}

		// Do we already have a server with this address?
		for _, existingAddr := range conf.DNS.Servers {
			if existingAddr.String() == address {
				return nil, errors.New("server already exists")
			}
		}

		var server types.MaybeAddrPort
		if err := server.UnmarshalText([]byte(address)); err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}

		// Add the new server.
		conf.DNS.Servers = append(conf.DNS.Servers, server)

		return conf, nil
	})
}
