//go:build windows

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

package shell

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/noisysockets/noisysockets/config/v1alpha1"
)

// When windows conpty support is added to creack/pty, this can be removed.
// See: https://github.com/creack/pty/pull/155
func Serve(ctx context.Context, logger *slog.Logger, conf *v1alpha1.Config) error {
	return fmt.Errorf("serve is not supported on Windows")
}
