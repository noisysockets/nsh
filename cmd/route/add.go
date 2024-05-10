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

package route

import (
	"fmt"
	"log/slog"

	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Add(logger *slog.Logger, configPath, destination, via string) error {
	return util.UpdateConfig(logger, configPath, func(conf *v1alpha1.Config) (*v1alpha1.Config, error) {
		// Do we already have a route with this destination?
		for _, routeConf := range conf.Routes {
			if routeConf.Destination == destination {
				return nil, fmt.Errorf("route already exists")
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
			return nil, fmt.Errorf("gateway peer not found")
		}

		if err := validate.CIDR(destination); err != nil {
			return nil, fmt.Errorf("invalid destination: %w", err)
		}

		// Add the new route.
		conf.Routes = append(conf.Routes, v1alpha1.RouteConfig{
			Destination: destination,
			Via:         via,
		})

		return conf, nil
	})
}
