// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package middleware

import "net/http"

// FrameOptionType represents the possible values for the X-Frame-Options header.
type FrameOptionType string

const (
	// FrameOptionDeny prevents the page from being rendered in a frame.
	FrameOptionDeny FrameOptionType = "DENY"
	// FrameOptionSameOrigin allows the page to be rendered in a frame on the same origin.
	FrameOptionSameOrigin FrameOptionType = "SAMEORIGIN"
	// FrameOptionAllowFrom allows the page to be rendered in a frame on the specified origin.
	FrameOptionAllowFrom FrameOptionType = "ALLOW-FROM"
)

// FrameOptions sets the X-Frame-Options header in the response.
func FrameOptions(option FrameOptionType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", string(option))

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}
