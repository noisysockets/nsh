// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"

	"github.com/noisysockets/network"
	"github.com/noisysockets/network/forwarder"
	"github.com/noisysockets/noisysockets"
)

var _ Service = (*RouterService)(nil)

// RouterService is a service that forwards packets from the source network to
// the destination network and vice versa.
type RouterService struct {
	logger      *slog.Logger
	dstNet      network.Network
	enableNAT64 bool
	nat64Prefix netip.Prefix
}

// Router returns a service that forwards packets from the source network to
// the destination network and vice versa.
func Router(logger *slog.Logger, dstNet network.Network, enableNAT64 bool, nat64Prefix netip.Prefix) *RouterService {
	return &RouterService{
		logger:      logger,
		dstNet:      dstNet,
		enableNAT64: enableNAT64,
		nat64Prefix: nat64Prefix,
	}
}

func (s *RouterService) Serve(ctx context.Context, net network.Network) error {
	s.logger.Info("Enabling packet forwarding")

	fwdConf := forwarder.ForwarderConfig{
		AllowedDestinations: []netip.Prefix{
			netip.MustParsePrefix("0.0.0.0/0"),
			netip.MustParsePrefix("::/0"),
		},
		EnableNAT64: &s.enableNAT64,
		NAT64Prefix: &s.nat64Prefix,
	}

	userspaceNet := net.(*noisysockets.NoisySocketsNetwork).UserspaceNetwork
	fwd, err := forwarder.New(ctx, s.logger, userspaceNet, s.dstNet, &fwdConf)
	if err != nil {
		return fmt.Errorf("failed to create packet forwarder: %w", err)
	}
	defer fwd.Close()

	if err := net.(*noisysockets.NoisySocketsNetwork).EnableForwarding(fwd); err != nil {
		return fmt.Errorf("failed to enable packet forwarding: %w", err)
	}

	<-ctx.Done()

	return nil
}
