// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"context"
	"net"
	"net/netip"

	"github.com/nextmn/ue-lite/internal/session"
	"github.com/nextmn/ue-lite/internal/tun"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

type RadioDaemon struct {
	Control   jsonapi.ControlURI
	Gnbs      []jsonapi.ControlURI
	Radio     *Radio
	PsMan     *session.PduSessionsManager
	UeRanAddr netip.AddrPort
	tunMan    *tun.TunManager
	closed    chan struct{}
}

func NewRadioDaemon(control jsonapi.ControlURI, gnbs []jsonapi.ControlURI, radio *Radio, psMan *session.PduSessionsManager, tunMan *tun.TunManager, ueRanAddr netip.AddrPort) *RadioDaemon {
	return &RadioDaemon{
		Control:   control,
		Gnbs:      gnbs,
		Radio:     radio,
		PsMan:     psMan,
		UeRanAddr: ueRanAddr,
		tunMan:    tunMan,
		closed:    make(chan struct{}),
	}
}

func (r *RadioDaemon) runDownlinkDaemon(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) error {
	if srv == nil {
		return ErrNilUdpConn
	}
	if ifacetun == nil {
		return ErrNilTunIface
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, tun.TUN_MTU)
			n, err := srv.Read(buf)
			if err != nil {
				return err
			}
			ifacetun.Write(buf[:n])
		}
	}
	return nil
}

func (r *RadioDaemon) runUplinkDaemon(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) error {
	if srv == nil {
		return ErrNilUdpConn
	}
	if ifacetun == nil {
		return ErrNilTunIface
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, tun.TUN_MTU)
			n, err := ifacetun.Read(buf)
			if err != nil {
				return err
			}

			// get UE IP Address
			if !waterutil.IsIPv4(buf[:n]) {
				return ErrUnsupportedPDUType
			}
			src, ok := netip.AddrFromSlice(waterutil.IPv4Source(buf[:n]).To4())
			if !ok {
				return ErrMalformedPDU
			}

			// get gNB linked to UE
			gnb, err := r.PsMan.LinkedGnb(src)
			if err != nil {
				return err
			}
			if err := r.Radio.Write(buf[:n], srv, gnb); err == nil {
				logrus.WithFields(
					logrus.Fields{
						"ip-addr": src,
					}).Trace("Packet forwarded")
			}
		}
	}
	return nil
}

func (r *RadioDaemon) Start(ctx context.Context) error {
	if err := r.Radio.Init(ctx); err != nil {
		return err
	}
	ifacetun := r.tunMan.OpenTun()
	defer func(ctx context.Context) {
		defer r.tunMan.CloseTun()
		select {
		case <-ctx.Done():
			close(r.closed)
			return
		}
	}(ctx)
	srv, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(r.UeRanAddr))
	if err != nil {
		return err
	}
	go func(ctx context.Context, srv *net.UDPConn) error {
		if srv == nil {
			return ErrNilUdpConn
		}
		select {
		case <-ctx.Done():
			srv.Close()
			return ctx.Err()
		}
		return nil
	}(ctx, srv)
	go func(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) {
		r.runDownlinkDaemon(ctx, srv, ifacetun)
	}(ctx, srv, ifacetun)
	go func(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) {
		r.runUplinkDaemon(ctx, srv, ifacetun)
	}(ctx, srv, ifacetun)

	for _, gnb := range r.Gnbs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := r.Radio.InitPeer(gnb); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RadioDaemon) WaitShutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.closed:
		return nil
	}
}
