// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"context"
	"fmt"
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
}

func NewRadioDaemon(control jsonapi.ControlURI, gnbs []jsonapi.ControlURI, radio *Radio, psMan *session.PduSessionsManager, tunMan *tun.TunManager, ueRanAddr netip.AddrPort) *RadioDaemon {
	return &RadioDaemon{
		Control:   control,
		Gnbs:      gnbs,
		Radio:     radio,
		PsMan:     psMan,
		UeRanAddr: ueRanAddr,
		tunMan:    tunMan,
	}
}

func (r *RadioDaemon) runDownlinkDaemon(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) error {
	if srv == nil {
		return fmt.Errorf("nil srv")
	}
	if ifacetun == nil {
		return fmt.Errorf("nil tun iface")
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
		return fmt.Errorf("nil srv")
	}
	if ifacetun == nil {
		return fmt.Errorf("nil tun iface")
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
				return fmt.Errorf("not an IPv4 packet")
			}
			src, ok := netip.AddrFromSlice(waterutil.IPv4Source(buf[:n]).To4())
			if !ok {
				return fmt.Errorf("error while retrieving ip addr")
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
					}).Trace("packet forwarded")
			}
		}
	}
	return nil
}

func (r *RadioDaemon) Start(ctx context.Context) error {
	ifacetun := r.tunMan.Tun
	srv, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(r.UeRanAddr))
	if err != nil {
		return err
	}
	go func(ctx context.Context, srv *net.UDPConn) error {
		if srv == nil {
			return fmt.Errorf("nil srv")
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
			if err := r.Radio.InitPeer(ctx, gnb); err != nil {
				return err
			}
		}
	}
	return nil
}
