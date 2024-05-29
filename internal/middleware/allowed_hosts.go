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
	"net/netip"
	"net/url"

	"github.com/miekg/dns"
)

// Prevent DNS rebinding attacks by validating the Host header.
func AllowedHosts(allowedOrigins []string) func(next http.Handler) http.Handler {
	// Parse the origins to get the hostnames.
	allowedHosts := make(map[string]bool)
	for _, origin := range allowedOrigins {
		parsedURL, err := url.Parse(origin)
		if err != nil {
			// It's alright, it fails safely.
			continue
		}

		hostname := parsedURL.Hostname()
		if _, err := netip.ParseAddr(hostname); err == nil {
			allowedHosts[hostname] = true
		} else {
			allowedHosts[dns.Fqdn(hostname)] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var allowedHost bool
			if _, err := netip.ParseAddr(r.Host); err == nil {
				allowedHost = allowedHosts[r.Host]
			} else {
				allowedHost = allowedHosts[dns.Fqdn(r.Host)]
			}

			if !allowedHost {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
