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
	"fmt"
	"log/slog"
	"time"

	stdnet "net"

	"github.com/miekg/dns"
	"github.com/noisysockets/network"
	"github.com/noisysockets/resolver"
)

// dnsService is an emdedded DNS service.
type dnsService struct {
	logger *slog.Logger
}

// DNS creates a new instance of the DNS service.
func DNS(logger *slog.Logger) *dnsService {
	return &dnsService{
		logger: logger,
	}
}

func (s *dnsService) Serve(ctx context.Context, net network.Network) error {
	domain, err := net.Domain()
	if err != nil {
		return fmt.Errorf("failed to get network domain: %w", err)
	}

	mux := dns.NewServeMux()

	s.logger.Info("Registering recursive DNS handler")

	// TODO: make this configurable.
	upstreamResolver := resolver.Default

	mux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		reply := &dns.Msg{}
		reply.SetReply(req)

		logger := s.logger.With(
			slog.String("zone", "."),
			slog.String("remoteAddr", w.RemoteAddr().String()),
			slog.Int("id", int(req.Id)))

		defer func() {
			if err := w.WriteMsg(reply); err != nil {
				logger.Error("Failed to write DNS response", slog.Any("error", err))
			}
		}()

		// Make sure the client is asking for a recursive query.
		if !req.RecursionDesired {
			logger.Warn("Non-recursive query")

			reply.Rcode = dns.RcodeRefused
			return
		}

		logger.Debug("Forwarding DNS question")

		for _, q := range req.Question {
			logger = logger.With(
				slog.String("name", q.Name),
				slog.String("qType", dns.TypeToString[q.Qtype]))

			logger.Debug("Received DNS question")

			addrs, err := upstreamResolver.LookupNetIP(ctx, "ip", q.Name)
			if err != nil {
				logger.Warn("Failed to lookup DNS question", slog.Any("error", err))
				reply.Rcode = dns.RcodeNameError
				return
			}

			var ipv4Addrs, ipv6Addrs []stdnet.IP
			for _, addr := range addrs {
				if addr.Unmap().Is4() {
					ipv4Addrs = append(ipv4Addrs, stdnet.IP(addr.Unmap().AsSlice()))
				} else {
					ipv6Addrs = append(ipv6Addrs, stdnet.IP(addr.AsSlice()))
				}
			}

			switch q.Qtype {
			case dns.TypeA:
				logger.Debug("Answering DNS question", slog.Int("answers", len(ipv4Addrs)))

				for _, addr := range ipv4Addrs {
					reply.Answer = append(reply.Answer, &dns.A{
						Hdr: dns.RR_Header{
							Name:   q.Name,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							// TODO: would be nice to get ttl from the upstream.
							Ttl: 300,
						},
						A: addr,
					})
				}
			case dns.TypeAAAA:
				logger.Debug("Answering DNS question", slog.Int("answers", len(ipv6Addrs)))

				for _, addr := range ipv6Addrs {
					reply.Answer = append(reply.Answer, &dns.AAAA{
						Hdr: dns.RR_Header{
							Name:   q.Name,
							Rrtype: dns.TypeAAAA,
							Class:  dns.ClassINET,
							// TODO: would be nice to get ttl from the upstream.
							Ttl: 300,
						},
						AAAA: addr,
					})
				}
			default:
				logger.Warn("Unsupported DNS query type")

				reply.Rcode = dns.RcodeNotImplemented
			}
		}
	})

	s.logger.Info("Registering authoritive DNS handler", slog.String("zone", domain))

	mux.HandleFunc(domain, func(w dns.ResponseWriter, req *dns.Msg) {
		reply := &dns.Msg{}
		reply.SetReply(req)

		logger := s.logger.With(
			slog.String("zone", domain),
			slog.String("remoteAddr", w.RemoteAddr().String()),
			slog.Int("id", int(req.Id)))

		defer func() {
			if err := w.WriteMsg(reply); err != nil {
				logger.Error("Failed to write DNS response", slog.Any("error", err))
			}
		}()

		for _, q := range req.Question {
			logger = logger.With(
				slog.String("name", q.Name),
				slog.String("qType", dns.TypeToString[q.Qtype]))

			logger.Debug("Received DNS question")

			addrs, err := net.LookupHost(q.Name)
			if err != nil {
				logger.Warn("Failed to lookup DNS question", slog.Any("error", err))
				reply.Rcode = dns.RcodeNameError
				return
			}

			var ipv4Addrs, ipv6Addrs []stdnet.IP
			for _, addr := range addrs {
				ip := stdnet.ParseIP(addr)
				if ip == nil {
					logger.Warn("Failed to parse IP address", slog.String("address", addr))
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
				logger.Debug("Answering DNS question", slog.Int("answers", len(ipv4Addrs)))

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
				logger.Debug("Answering DNS question", slog.Int("answers", len(ipv6Addrs)))

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
				logger.Warn("Unsupported DNS query type")

				reply.Rcode = dns.RcodeNotImplemented
			}
		}
	})

	s.logger.Debug("Binding to TCP port")

	lis, err := net.Listen("tcp", ":53")
	if err != nil {
		return fmt.Errorf("failed to listen on TCP port: %w", err)
	}
	defer lis.Close()

	s.logger.Debug("Binding to UDP port")

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

		s.logger.Debug("Shutting down DNS server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.ShutdownContext(shutdownCtx); err != nil {
			s.logger.Error("Failed to shutdown DNS server", slog.Any("error", err))
		}
	}()

	s.logger.Info("Starting DNS server", slog.String("address", lis.Addr().String()))

	if err := srv.ActivateAndServe(); err != nil {
		return fmt.Errorf("failed to start DNS server: %w", err)
	}

	return nil
}
