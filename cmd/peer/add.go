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

package peer

import (
	"fmt"

	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/noisysockets/types"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Add(configPath, name, publicKey, endpoint string, ips []string) error {
	return util.UpdateConfig(configPath, func(conf *v1alpha1.Config) (*v1alpha1.Config, error) {
		// Do we already have a peer with this name or public key?
		for _, peerConf := range conf.Peers {
			if peerConf.Name == name || peerConf.PublicKey == publicKey {
				return nil, fmt.Errorf("peer already exists")
			}
		}

		// Validate the public key.
		var pk types.NoisePublicKey
		if err := pk.FromString(publicKey); err != nil {
			return nil, fmt.Errorf("invalid public key: %w", err)
		}

		if err := validate.Endpoint(endpoint); err != nil {
			return nil, fmt.Errorf("invalid endpoint: %w", err)
		}

		if err := validate.IPs(ips); err != nil {
			return nil, fmt.Errorf("invalid IP address: %w", err)
		}

		// Add the new peer.
		conf.Peers = append(conf.Peers, v1alpha1.PeerConfig{
			Name:      name,
			PublicKey: publicKey,
			Endpoint:  endpoint,
			IPs:       ips,
		})

		return conf, nil
	})
}
