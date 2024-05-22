//go:build windows

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
	"errors"
	"log/slog"

	"github.com/noisysockets/network"
)

// shellService is a remote shell service.
type shellService struct{}

// Shell returns a new remote shell service.
func Shell(_ *slog.Logger) *shellService {
	return &shellService{}
}

// When windows conpty support is added to creack/pty, this can be removed.
// See: https://github.com/creack/pty/pull/155
func (s *shellService) Serve(_ context.Context, _ network.Network) error {
	return errors.New("shell service is not supported on Windows")
}
