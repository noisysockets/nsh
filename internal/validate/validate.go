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
	"net/netip"
)

// IPs validates a list of IP addresses.
func IPs(ips []string) error {
	for _, ip := range ips {
		if err := IP(ip); err != nil {
			return err
		}
	}
	return nil
}

// IP validates an IP address string.
func IP(ip string) error {
	_, err := netip.ParseAddr(ip)
	if err != nil {
		return fmt.Errorf("invalid IP address %q: %w", ip, err)
	}
	return nil
}

// Endpoint validates an endpoint string.
func Endpoint(endpoint string) error {
	_, _, err := stdnet.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}
	return nil
}

// CIDR validates a network CIDR string.
func CIDR(cidr string) error {
	_, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	return nil
}
