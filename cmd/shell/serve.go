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
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/noisysockets/noisysockets"
	"github.com/noisysockets/noisysockets/config/v1alpha1"
	"github.com/noisysockets/nsh/web"
	"github.com/noisysockets/shell"
	"github.com/rs/cors"
)

func Serve(ctx context.Context, logger *slog.Logger, conf *v1alpha1.Config) error {
	logger.Debug("Opening WireGuard network")

	net, err := noisysockets.NewNetwork(logger, conf)
	if err != nil {
		return fmt.Errorf("failed to open WireGuard network: %w", err)
	}
	defer net.Close()

	logger.Debug("Binding to http port")

	lis, err := net.Listen("tcp", ":80")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer lis.Close()

	// The IP address and port that the listener is bound to.
	lisAddrPort := netip.MustParseAddrPort(lis.Addr().String())
	allowedOrigins := []string{
		fmt.Sprintf("http://%s", lisAddrPort.Addr()),
		fmt.Sprintf("http://%s", lisAddrPort.String()),
	}

	// The hostname of the shell server peer.
	hostname, err := net.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	if hostname != "" {
		allowedOrigins = append(allowedOrigins,
			fmt.Sprintf("http://%s", hostname),
			fmt.Sprintf("http://%s:%d", hostname, lisAddrPort.Port()))
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: corsHandler.OriginAllowed,
	}

	mux := http.NewServeMux()

	mux.Handle("/", web.Handler())

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		logger := logger.With(slog.String("remote_addr", r.RemoteAddr))

		logger.Info("Handling connection")

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("Error upgrading connection", slog.Any("error", err))
			return
		}

		h, err := shell.NewHandler(ctx, logger, ws)
		if err != nil {
			logger.Error("Failed to create shell handler", slog.Any("error", err))
			return
		}
		defer h.Close()

		// Wait for the handler to complete (eg. shell process exits).
		if err := h.Wait(); err != nil {
			logger.Error("Error handling connection", slog.Any("error", err))
		} else {
			logger.Info("Finished handling connection")
		}
	})

	srv := &http.Server{
		Handler: mux,
	}

	// Capture the signal to close the listener
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig

		if err := srv.Close(); err != nil {
			logger.Error("Failed to close server", slog.Any("error", err))
		}
	}()

	urlStr := fmt.Sprintf("http://%s", lisAddrPort.String())
	if hostname != "" {
		urlStr = fmt.Sprintf("http://%s:%d", hostname, lisAddrPort.Port())
	}

	logger.Info("Listening for shell connections on WireGuard network", slog.String("url", urlStr))

	// Serve connections.
	if err := srv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
