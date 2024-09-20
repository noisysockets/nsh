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
	"net/netip"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha3"
	"github.com/noisysockets/nsh/internal/util"
)

func Add(configPath, destination, via string) error {
	return util.UpdateConfig(configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		// Do we already have a route with this destination?
		for _, routeConf := range conf.Routes {
			if routeConf.Destination.String() == destination {
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
			return nil, errors.New("router peer not found")
		}

		destinationPrefix, err := netip.ParsePrefix(destination)
		if err != nil {
			return nil, fmt.Errorf("invalid destination: %w", err)
		}

		// Add the new route.
		conf.Routes = append(conf.Routes, latestconfig.RouteConfig{
			Destination: destinationPrefix,
			Via:         via,
		})

		return conf, nil
	})
}
