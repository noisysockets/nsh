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

package config

import (
	"context"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/noisysockets/types"
	"gopkg.in/yaml.v3"
)

// Show queries the config with the provided jq syntax query and prints the
// result to stdout as YAML.
func Show(ctx context.Context, conf *v1alpha1.Config, queryStr string) error {
	query, err := gojq.Parse(queryStr)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	// Add a custom function to the jq query to convert a private key to a public key.
	code, err := gojq.Compile(query, gojq.WithFunction("public", 1, 1, func(_ any, xs []any) any {
		var privateKey types.NoisePrivateKey
		if err := privateKey.UnmarshalText([]byte(fmt.Sprintf("%v", xs[0]))); err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}

		return privateKey.Public().String()
	}))
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
