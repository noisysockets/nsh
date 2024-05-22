// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/noisysockets/noisysockets"
	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
	"github.com/noisysockets/nsh/internal/service"
	"golang.org/x/sync/errgroup"
)

func Serve(ctx context.Context, logger *slog.Logger, conf *latestconfig.Config, services []service.Service) error {
	logger.Debug("Opening WireGuard network")

	net, err := noisysockets.OpenNetwork(logger, conf)
	if err != nil {
		return fmt.Errorf("failed to open WireGuard network: %w", err)
	}
	defer net.Close()

	g, ctx := errgroup.WithContext(ctx)

	// Capture the signal to close the listener
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	g.Go(func() error {
		<-sig

		logger.Debug("Received signal, shutting down")

		return context.Canceled
	})

	for _, s := range services {
		g.Go(func() error {
			return s.Serve(ctx, net)
		})
	}

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
