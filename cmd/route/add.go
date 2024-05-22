// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package route

import (
	"errors"
	"fmt"
	"log/slog"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Add(logger *slog.Logger, configPath, destination, via string) error {
	return util.UpdateConfig(logger, configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		// Do we already have a route with this destination?
		for _, routeConf := range conf.Routes {
			if routeConf.Destination == destination {
				return nil, errors.New("route already exists")
			}
		}

		var found bool
		for _, peerConf := range conf.Peers {
			if peerConf.Name == via || peerConf.PublicKey == via {
				found = true
				break
			}
		}

		if !found {
			return nil, errors.New("gateway peer not found")
		}

		if err := validate.CIDR(destination); err != nil {
			return nil, fmt.Errorf("invalid destination: %w", err)
		}

		// Add the new route.
		conf.Routes = append(conf.Routes, latestconfig.RouteConfig{
			Destination: destination,
			Via:         via,
		})

		return conf, nil
	})
}
