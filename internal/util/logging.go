// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util

import (
	"log/slog"
	"strings"
)

// LevelFlag is a urfave/cli compatible flag for setting the log verbosity level.
type LevelFlag slog.Level

func FromSlogLevel(l slog.Level) *LevelFlag {
	f := LevelFlag(l)
	return &f
}

func (f *LevelFlag) Set(value string) error {
	return (*slog.Level)(f).UnmarshalText([]byte(strings.ToUpper(value)))
}

func (f *LevelFlag) String() string {
	return (*slog.Level)(f).String()
}
