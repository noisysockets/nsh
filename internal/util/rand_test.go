// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util_test

import (
	"testing"

	"github.com/noisysockets/nsh/internal/util"
	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	values := make([]int, 100)
	for i := 0; i < 100; i++ {
		values[i] = util.RandomInt(0, 100)
	}

	// Make sure the values are within the expected range.
	for _, v := range values {
		require.True(t, v >= 0 && v < 100)
	}

	// Check the mean of the values.
	sum := 0
	for _, v := range values {
		sum += v
	}

	mean := sum / 100
	require.InDelta(t, 50, mean, 10)
}
