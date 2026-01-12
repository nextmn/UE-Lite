// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"context"
	"net"
	"net/netip"

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
	UeRanAddr netip.AddrPort
	closed    chan struct{}
}

func NewRadioDaemon(control jsonapi.ControlURI, gnbs []jsonapi.ControlURI, radio *Radio, ueRanAddr netip.AddrPort) *RadioDaemon {
	return &RadioDaemon{
		Control:   control,
		Gnbs:      gnbs,
		Radio:     radio,
		UeRanAddr: ueRanAddr,
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
			if err := r.handleUplinkPDU(srv, ifacetun); err != nil {
				logrus.WithError(err).Trace("Packet dropped")
			}
		}
	}
}

func (r *RadioDaemon) handleUplinkPDU(srv *net.UDPConn, ifacetun *water.Interface) error {
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

	if err := r.Radio.Write(buf[:n], srv, src); err == nil {
		logrus.WithFields(
			logrus.Fields{
				"ip-addr": src,
			}).Trace("Packet forwarded")
	}
	return err

}

func (r *RadioDaemon) Start(ctx context.Context) error {
	if err := r.Radio.InitContext(ctx); err != nil {
		return err
	}
	ifacetun := r.Radio.Tun.OpenTun()
	go func(ctx context.Context) error {
		defer r.Radio.Tun.CloseTun()
		<-ctx.Done()
		close(r.closed)
		return ctx.Err()
	}(ctx)
	srv, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(r.UeRanAddr))
	if err != nil {
		return err
	}
	go func(ctx context.Context, srv *net.UDPConn) error {
		if srv == nil {
			return ErrNilUdpConn
		}
		<-ctx.Done()
		srv.Close()
		return ctx.Err()
	}(ctx, srv)
	go func(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) {
		if err := r.runDownlinkDaemon(ctx, srv, ifacetun); err != nil {
			logrus.WithError(err).Error("Radio Downlink Daemon stopped")
		}
	}(ctx, srv, ifacetun)
	go func(ctx context.Context, srv *net.UDPConn, ifacetun *water.Interface) {
		if err := r.runUplinkDaemon(ctx, srv, ifacetun); err != nil {
			logrus.WithError(err).Error("Radio Uplink Daemon stopped")
		}
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
