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
	"crypto/rand"
	"math/big"
)

// RandomInt returns a random integer min <= n < max.
func RandomInt(min, max int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		panic(err)
	}
	return min + int(nBig.Int64())
}
