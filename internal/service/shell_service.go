//go:build !windows

// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdnet "net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
	"github.com/noisysockets/network"
	"github.com/noisysockets/nsh/internal/middleware"
	"github.com/noisysockets/nsh/web"
	resolverutil "github.com/noisysockets/resolver/util"
	"github.com/noisysockets/shell"
	"github.com/rs/cors"
)

// shellService is a remote shell service.
type shellService struct {
	logger *slog.Logger
}

// Shell returns a new remote shell service.
func Shell(logger *slog.Logger) *shellService {
	return &shellService{
		logger: logger,
	}
}

// Serve starts the shell service.
func (s *shellService) Serve(ctx context.Context, net network.Network) error {
	s.logger.Debug("Binding to http port")

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
		domain, err := net.Domain()
		if err != nil {
			return fmt.Errorf("failed to get network domain: %w", err)
		}

		names := []string{hostname, strings.TrimSuffix(resolverutil.Join(hostname, domain), ".")}

		for _, name := range names {
			allowedOrigins = append(allowedOrigins,
				fmt.Sprintf("http://%s", name),
				fmt.Sprintf("http://%s:%d", name, lisAddrPort.Port()),
				fmt.Sprintf("http://%s", dns.Fqdn(name)),
				fmt.Sprintf("http://%s:%d", dns.Fqdn(name), lisAddrPort.Port()),
			)
		}
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
	})

	upgrader := websocket.Upgrader{
		CheckOrigin: corsHandler.OriginAllowed,
	}

	mux := http.NewServeMux()

	mux.Handle("/", web.Handler())
	mux.Handle("/shell/", http.StripPrefix("/shell", web.Handler()))

	mux.HandleFunc("/shell/ws", func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(slog.String("remoteAddr", r.RemoteAddr))

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

	middlewares := []middleware.Middleware{
		middleware.Recover(s.logger),
		middleware.FrameOptions(middleware.FrameOptionDeny),
		middleware.ContentSecurityPolicy,
		corsHandler.Handler,
	}

	srv := &http.Server{
		BaseContext: func(_ stdnet.Listener) context.Context {
			return ctx
		},
		Handler: middleware.Chain(middlewares...)(mux),
	}

	go func() {
		<-ctx.Done()

		if err := srv.Close(); err != nil {
			s.logger.Error("Failed to close server", slog.Any("error", err))
		}
	}()

	urlStr := fmt.Sprintf("http://%s/shell/", lisAddrPort.String())
	if hostname != "" {
		urlStr = fmt.Sprintf("http://%s:%d/shell/", hostname, lisAddrPort.Port())
	}

	s.logger.Info("Listening for shell connections on WireGuard network", slog.String("url", urlStr))

	if err := srv.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
