// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package dns

import (
	"context"
	"fmt"
	"log/slog"
	stdnet "net"
	"time"

	"github.com/miekg/dns"
	"github.com/noisysockets/noisysockets"
	latestconfig "github.com/noisysockets/noisysockets/config/v1alpha2"
)

// TODO: add support for recursive DNS queries.
func Serve(ctx context.Context, logger *slog.Logger, conf *latestconfig.Config) error {
	logger.Debug("Opening WireGuard network")

	net, err := noisysockets.OpenNetwork(logger, conf)
	if err != nil {
		return fmt.Errorf("failed to open WireGuard network: %w", err)
	}
	defer net.Close()

	mux := dns.NewServeMux()

	mux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		logger.Debug("Received DNS request", slog.Any("request", req))

		reply := &dns.Msg{}
		reply.SetReply(req)

		defer func() {
			if err := w.WriteMsg(reply); err != nil {
				logger.Error("Failed to write DNS response", slog.Any("error", err))
			}
		}()

		for _, q := range req.Question {
			logger.Debug("Received DNS question", slog.Any("question", q))

			addrs, err := net.LookupHost(q.Name)
			if err != nil {
				logger.Error("Failed to lookup DNS question", slog.Any("error", err))
				reply.Rcode = dns.RcodeNameError
				return
			}

			var ipv4Addrs, ipv6Addrs []stdnet.IP
			for _, addr := range addrs {
				ip := stdnet.ParseIP(addr)
				if ip == nil {
					logger.Debug("Failed to parse IP address", slog.String("address", addr))
					continue
				}

				if ip.To4() != nil {
					ipv4Addrs = append(ipv4Addrs, ip)
				} else {
					ipv6Addrs = append(ipv6Addrs, ip)
				}
			}

			switch q.Qtype {
			case dns.TypeA:
				logger.Debug("Answering DNS question", slog.Any("question", q))

				for _, addr := range ipv4Addrs {
					reply.Answer = append(reply.Answer, &dns.A{
						Hdr: dns.RR_Header{
							Name:   q.Name,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    60,
						},
						A: addr,
					})
				}
			case dns.TypeAAAA:
				logger.Debug("Answering DNS question", slog.Any("question", q))

				for _, addr := range ipv6Addrs {
					reply.Answer = append(reply.Answer, &dns.AAAA{
						Hdr: dns.RR_Header{
							Name:   q.Name,
							Rrtype: dns.TypeAAAA,
							Class:  dns.ClassINET,
							Ttl:    60,
						},
						AAAA: addr,
					})
				}
			default:
				logger.Warn("Unsupported DNS query type", slog.Int("queryType", int(q.Qtype)))
				reply.Rcode = dns.RcodeNotImplemented
			}
		}
	})

	logger.Debug("Binding to TCP port")

	lis, err := net.Listen("tcp", ":53")
	if err != nil {
		return fmt.Errorf("failed to listen on TCP port: %w", err)
	}
	defer lis.Close()

	logger.Debug("Binding to UDP port")

	pc, err := net.ListenPacket("udp", ":53")
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port: %w", err)
	}
	defer pc.Close()

	srv := &dns.Server{
		Handler:    mux,
		Listener:   lis,
		PacketConn: pc,
	}

	go func() {
		<-ctx.Done()

		logger.Debug("Shutting down DNS server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.ShutdownContext(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown DNS server", slog.Any("error", err))
		}
	}()

	logger.Info("Starting DNS server", slog.String("address", lis.Addr().String()))

	if err := srv.ActivateAndServe(); err != nil {
		return fmt.Errorf("failed to start DNS server: %w", err)
	}

	return nil
}
