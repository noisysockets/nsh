//go:build !windows

// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package shell

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func listenForWindowChangeEvents(ctx context.Context) (<-chan os.Signal, error) {
	windowChange := make(chan os.Signal, 1)
	go func() {
		<-ctx.Done()
		close(windowChange)
	}()
	signal.Notify(windowChange, syscall.SIGWINCH)
	return windowChange, nil
}
