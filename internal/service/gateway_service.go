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
	"log/slog"

	"github.com/noisysockets/network"
	"github.com/noisysockets/network/forwarder"
	"github.com/noisysockets/noisysockets"
)

var _ Service = (*GatewayService)(nil)

// GatewayService is a service that forwards packets from the WireGuard network
// to the destination network.
type GatewayService struct {
	logger         *slog.Logger
	destinationNet network.Network
}

// Gateway returns a new gateway service.
func Gateway(logger *slog.Logger, destinationNet network.Network) *GatewayService {
	return &GatewayService{
		logger:         logger,
		destinationNet: destinationNet,
	}
}

func (s *GatewayService) Serve(ctx context.Context, net network.Network) error {
	fwd := forwarder.New(ctx, s.logger, s.destinationNet, nil)
	defer fwd.Close()

	net.(*noisysockets.NoisySocketsNetwork).Network.(*network.UserspaceNetwork).EnableForwarding(fwd)

	<-ctx.Done()

	return nil
}
