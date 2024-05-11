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
	"errors"
	"fmt"
	stdio "io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/noisysockets/noisysockets"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/nsh/internal/util"
	"github.com/noisysockets/shell"
	"github.com/noisysockets/shell/env"
	"github.com/noisysockets/shell/io"
	"golang.org/x/sync/errgroup"
	"golang.org/x/term"
)

func Connect(ctx context.Context, logger *slog.Logger, conf *v1alpha1.Config, address string) error {
	logger.Debug("Opening WireGuard network")

	net, err := noisysockets.OpenNetwork(logger, conf)
	if err != nil {
		return fmt.Errorf("failed to open WireGuard network: %w", err)
	}
	defer net.Close()

	logger.Debug("Connecting to server", slog.String("address", address))

	dialer := &websocket.Dialer{
		NetDial:          net.Dial,
		NetDialContext:   net.DialContext,
		HandshakeTimeout: 45 * time.Second,
	}

	urlStr := fmt.Sprintf("ws://%s/ws", address)
	ws, _, err := dialer.Dial(urlStr, http.Header{
		"Origin": []string{fmt.Sprintf("http://%s", address)},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	logger.Debug("Creating client")

	c, err := shell.NewClient(ctx, logger, ws)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Set the terminal to raw mode.
	origTermState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), origTermState); err != nil {
			logger.Warn("Failed to restore terminal state", slog.Any("error", err))
		}
	}()

	// Create a pipe so we can make stdin non-blocking.
	input, inputWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer input.Close()

	// Pump stdin into the pipe (blocking in a seperate goroutine).
	// We can't use errgroup here as context cancellation is not available.
	go func() {
		defer inputWriter.Close()

		if _, err := stdio.Copy(inputWriter, os.Stdin); err != nil && !errors.Is(err, io.ErrClosedPipe) {
			logger.Error("Failed to read from stdin", slog.Any("error", err))
		}
	}()

	if err := io.SetNonblock(input); err != nil {
		return fmt.Errorf("failed to set stdin file descriptor to non-blocking: %w", err)
	}

	output := io.NopDeadlineWriter(os.Stdout)

	ctx, cancel := context.WithCancel(ctx)

	g, ctx := errgroup.WithContext(ctx)

	// Close the client on cancellation of the error group.
	g.Go(func() error {
		<-ctx.Done()

		return c.Close()
	})

	// Catch unexpected client errors (eg. dropped connections).
	g.Go(func() error {
		if err := c.Wait(); err != nil {
			return err
		}

		return context.Canceled
	})

	// Handle the shell exit.
	onExit := func(exitStatus int) error {
		if exitStatus != 0 {
			logger.Debug("Shell exited with non-zero status", slog.Int("exit_status", exitStatus))

			g.Go(func() error {
				return util.ExitError(exitStatus)
			})
		} else {
			cancel()
		}

		return nil
	}

	logger.Debug("Opening shell")

	// Get the initial terminal dimensions.
	columns, rows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to get terminal size: %w", err)
	}

	// Request a new terminal.
	env := env.FilterSafe(os.Environ())
	if err := c.OpenTerminal(columns, rows, env, input, output, onExit); err != nil {
		return fmt.Errorf("failed to open shell: %w", err)
	}

	// Listen for window change events.
	windowChangeEv, err := listenForWindowChangeEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to listen for window change events: %w", err)
	}

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-windowChangeEv:
				columns, rows, err := term.GetSize(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to get terminal size: %w", err)
				}

				if err := c.ResizeTerminal(columns, rows); err != nil {
					return fmt.Errorf("failed to resize PTY: %w", err)
				}

				logger.Debug("Resized PTY", slog.Int("columns", columns), slog.Int("rows", rows))
			}
		}
	})

	// Wait for the terminal to exit.
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
