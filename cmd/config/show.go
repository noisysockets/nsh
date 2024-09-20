// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package config

import (
	"context"
	"fmt"
	"net/netip"
	"os"

	"github.com/itchyny/gojq"
	configtypes "github.com/noisysockets/noisysockets/config/types"
	"github.com/noisysockets/noisysockets/types"
	"gopkg.in/yaml.v3"
)

// Show queries the config with the provided jq syntax query and prints the
// result to stdout as YAML.
func Show(ctx context.Context, conf configtypes.Config, queryStr string) error {
	query, err := gojq.Parse(queryStr)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	funcs := []gojq.CompilerOption{
		// Get the public key from a private key.
		gojq.WithFunction("public", 1, 1, func(_ any, xs []any) any {
			var privateKey types.NoisePrivateKey
			if err := privateKey.UnmarshalText([]byte(fmt.Sprintf("%v", xs[0]))); err != nil {
				return fmt.Errorf("failed to parse private key: %w", err)
			}

			return privateKey.Public().String()
		}),
		// Get the next address in a subnet.
		gojq.WithFunction("next", 1, 1, func(_ any, xs []any) any {
			addr, err := netip.ParseAddr(fmt.Sprintf("%v", xs[0]))
			if err != nil {
				return fmt.Errorf("failed to parse IP address: %w", err)
			}

			return addr.Next().String()
		}),
	}

	code, err := gojq.Compile(query, funcs...)
	if err != nil {
		return fmt.Errorf("failed to compile query: %w", err)
	}

	// Go to yaml and back so we can get a generic map to pass to gojq.
	yamlBytes, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	m := make(map[string]any)
	if err := yaml.Unmarshal(yamlBytes, &m); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	enc := yaml.NewEncoder(os.Stdout)
	defer enc.Close()

	iter := code.RunWithContext(ctx, m)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}

			return fmt.Errorf("failed to execute query: %w", err)
		}

		if err := enc.Encode(v); err != nil {
			return fmt.Errorf("failed to encode result: %w", err)
		}
	}

	return nil
}
