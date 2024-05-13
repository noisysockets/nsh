// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package peer

import (
	"fmt"
	"log/slog"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/nsh/internal/util"
)

func Remove(logger *slog.Logger, configPath, nameOrPublicKey string) error {
	return util.UpdateConfig(logger, configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		for i, peerConf := range conf.Peers {
			if peerConf.Name == nameOrPublicKey || peerConf.PublicKey == nameOrPublicKey {
				conf.Peers = append(conf.Peers[:i], conf.Peers[i+1:]...)
				return conf, nil
			}
		}

		return nil, fmt.Errorf("peer %q not found", nameOrPublicKey)
	})
}
