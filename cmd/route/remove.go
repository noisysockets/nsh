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

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha3"
	"github.com/noisysockets/nsh/internal/util"
)

func Remove(configPath, destination string) error {
	return util.UpdateConfig(configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		for i, routeConf := range conf.Routes {
			if routeConf.Destination.String() == destination {
				conf.Routes = append(conf.Routes[:i], conf.Routes[i+1:]...)
				return conf, nil
			}
		}

		return nil, errors.New("route not found")
	})
}
