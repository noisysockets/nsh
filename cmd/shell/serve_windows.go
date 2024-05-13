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
	"fmt"
	"log/slog"

	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
)

// When windows conpty support is added to creack/pty, this can be removed.
// See: https://github.com/creack/pty/pull/155
func Serve(_ context.Context, _ *slog.Logger, _ *latestconfig.Config) error {
	return fmt.Errorf("serve is not supported on Windows")
}
