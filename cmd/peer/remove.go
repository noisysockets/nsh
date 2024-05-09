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
	"github.com/noisysockets/nsh/internal/util"
)

func Remove(configPath, nameOrPublicKey string) error {
	return util.UpdateConfig(configPath, func(conf *v1alpha1.Config) (*v1alpha1.Config, error) {
		for i, peerConf := range conf.Peers {
			if peerConf.Name == nameOrPublicKey || peerConf.PublicKey == nameOrPublicKey {
				conf.Peers = append(conf.Peers[:i], conf.Peers[i+1:]...)
				return conf, nil
			}
		}

		return nil, fmt.Errorf("peer %q not found", nameOrPublicKey)
	})
}
