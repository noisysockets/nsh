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
	"errors"
	"fmt"
	"net/netip"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha3"
	"github.com/noisysockets/noisysockets/types"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/nsh/internal/validate"
)

func Add(configPath, name, publicKey, endpoint string, ips []string) error {
	return util.UpdateConfig(configPath, func(conf *latestconfig.Config) (*latestconfig.Config, error) {
		// Do we already have a peer with this name or public key?
		for _, peerConf := range conf.Peers {
			if peerConf.Name == name || peerConf.PublicKey == publicKey {
				return nil, errors.New("peer already exists")
			}
		}

		// Validate the public key.
		var pk types.NoisePublicKey
		if err := pk.UnmarshalText([]byte(publicKey)); err != nil {
			return nil, fmt.Errorf("invalid public key: %w", err)
		}

		var addrs []netip.Addr
		for _, ip := range ips {
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, fmt.Errorf("invalid IP address: %w", err)
			}

			addrs = append(addrs, addr)
		}

		if endpoint != "" {
			if err := validate.Endpoint(endpoint); err != nil {
				return nil, fmt.Errorf("invalid endpoint: %w", err)
			}
		}

		// Add the new peer.
		conf.Peers = append(conf.Peers, latestconfig.PeerConfig{
			Name:      name,
			PublicKey: publicKey,
			Endpoint:  endpoint,
			IPs:       addrs,
		})

		return conf, nil
	})
}
