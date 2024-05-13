// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util

import "fmt"

// ExitError is an error type that represents a shell exit status.
// This will be returned when the parent process exits.
type ExitError int

func (e ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e)
}
