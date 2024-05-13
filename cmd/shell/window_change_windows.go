//go:build windows

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
	"time"

	"golang.org/x/term"
)

// There really isn't a good way to listen for window change events on Windows,
// so poll window size every second or so.
// See: https://github.com/microsoft/terminal/issues/305
func listenForWindowChangeEvents(ctx context.Context) (<-chan os.Signal, error) {
	prevColumns, prevRows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)

	windowChange := make(chan os.Signal, 1)
	go func() {
		defer close(windowChange)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				columns, rows, err := term.GetSize(int(os.Stdin.Fd()))
				if err != nil {
					return
				}

				if columns != prevColumns || rows != prevRows {
					prevColumns = columns
					prevRows = rows

					windowChange <- os.Interrupt
				}
			}
		}
	}()

	return windowChange, nil
}
