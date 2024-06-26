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
	"net/netip"
	"strings"
	"time"

	stdnet "net"

	"github.com/miekg/dns"
	"github.com/noisysockets/network"
	"github.com/noisysockets/resolver"
	"golang.org/x/sync/errgroup"
)

var _ Service = (*DNSService)(nil)

// DNSService is a DNS service that provides recursive and authoritative DNS resolution.
type DNSService struct {
	logger      *slog.Logger
	enableNAT64 bool
	nat64Prefix netip.Prefix
}

// DNS returns a new DNS service.
func DNS(logger *slog.Logger, enableNAT64 bool, nat64Prefix netip.Prefix) *DNSService {
	return &DNSService{
		logger:      logger,
		enableNAT64: enableNAT64,
		nat64Prefix: nat64Prefix,
	}
}

func (s *DNSService) Serve(ctx context.Context, net network.Network) error {
	domain, err := net.Domain()
	if err != nil {
		return fmt.Errorf("failed to get network domain: %w", err)
	}

	mux := dns.NewServeMux()

	s.logger.Info("Registering recursive DNS handler", slog.String("zone", "."))

	// TODO: make configurable.
	upstreamResolver, err := resolver.System(nil)
	if err != nil {
		return fmt.Errorf("failed to get system resolver: %w", err)
	}

	if s.enableNAT64 {
		s.logger.Info("Enabling DNS64", slog.String("prefix", s.nat64Prefix.String()))

		upstreamResolver = resolver.DNS64(upstreamResolver, &resolver.DNS64ResolverConfig{
			Prefix: &s.nat64Prefix,
		})
	}

	mux.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		reply := &dns.Msg{}
		reply.SetReply(req)
		reply.RecursionAvailable = true

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

		logger.Info("Recursively resolving DNS query")

		for _, q := range req.Question {
			logger = logger.With(
				slog.String("name", q.Name),
				slog.String("qType", dns.TypeToString[q.Qtype]))

			logger.Debug("Received DNS question")

			addrs, err := upstreamResolver.LookupNetIP(ctx, "ip", q.Name)
			if err != nil {
				if strings.Contains(err.Error(), resolver.ErrNoSuchHost.Error()) {
					reply.Rcode = dns.RcodeNameError
					return
				}

				logger.Warn("Failed to lookup DNS question", slog.Any("error", err))
				reply.Rcode = dns.RcodeServerFailure
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
							// TODO: would be nice to get TTL from the upstream.
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
							// TODO: would be nice to get TTL from the upstream.
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
		reply.Authoritative = true
		reply.RecursionAvailable = true

		logger := s.logger.With(
			slog.String("zone", domain),
			slog.String("remoteAddr", w.RemoteAddr().String()),
			slog.Int("id", int(req.Id)))

		logger.Info("Resolving DNS question")

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
				if strings.Contains(err.Error(), resolver.ErrNoSuchHost.Error()) {
					reply.Rcode = dns.RcodeNameError
					return
				}

				logger.Warn("Failed to lookup DNS question", slog.Any("error", err))
				reply.Rcode = dns.RcodeServerFailure
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

	pc, err := net.ListenPacket("udp", ":53")
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port: %w", err)
	}
	defer pc.Close()

	// For UDP queries.
	udpServer := &dns.Server{
		Handler:    mux,
		PacketConn: pc,
	}

	lis, err := net.Listen("tcp", ":53")
	if err != nil {
		return fmt.Errorf("failed to listen on TCP port: %w", err)
	}
	defer lis.Close()

	// For TCP queries.
	tcpServer := &dns.Server{
		Handler:  mux,
		Listener: lis,
	}

	g, ctx := errgroup.WithContext(ctx)

	// We have to use multiple server instances as we can't serve both UDP and TCP
	// at the same time on the one server instance.
	for _, srv := range []*dns.Server{udpServer, tcpServer} {
		srv := srv

		g.Go(func() error {
			g.Go(func() error {
				<-ctx.Done()

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := srv.ShutdownContext(shutdownCtx); err != nil {
					return err
				}

				return nil
			})

			return srv.ActivateAndServe()
		})
	}

	s.logger.Info("Listening for DNS queries", slog.String("address", lis.Addr().String()))

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to serve DNS: %w", err)
	}

	return nil
}
