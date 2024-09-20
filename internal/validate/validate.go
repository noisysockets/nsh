// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package validate

import (
	"fmt"
	stdnet "net"
)

// Endpoint validates an endpoint string.
func Endpoint(endpoint string) error {
	_, _, err := stdnet.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}
	return nil
}
