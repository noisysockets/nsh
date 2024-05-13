// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package middleware

import (
	"net/http"
)

// ContentSecurityPolicy sets a locked down Content-Security-Policy header in
// the response.
func ContentSecurityPolicy(next http.Handler) http.Handler {
	// TODO: figure out how to implement a strict CSP policy.
	// See: https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html
	// See: https://csp.withgoogle.com/docs/strict-csp.html
	policy := "default-src 'none'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'; form-action 'self';"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", policy)

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
