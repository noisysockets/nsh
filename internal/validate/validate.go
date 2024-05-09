/*
 * Copyright 2024 Damian Peckett <damian@pecke.tt>
 *
 * Licensed under the Noisy Sockets Source License 1.0 (NSSL-1.0); you may not
 * use this file except in compliance with the License. You may obtain a copy
 * of the License at
 *
 * https://github.com/noisysockets/nsh/blob/main/LICENSE
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations
 * under the License.
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
		if _, err := netip.ParseAddr(ip); err != nil {
			return fmt.Errorf("invalid IP address %q: %w", ip, err)
		}
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
