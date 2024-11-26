// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/nextmn/ue-lite/internal/config"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/songgao/water"
)

type RadioDaemon struct {
	Control            jsonapi.ControlURI
	Gnbs               []jsonapi.ControlURI
	ReqPS              []config.PDUSession
	Radio              *Radio
	PduSessions        *PduSessions
	PduSessionsManager *PduSessionsManager
	UeRanAddr          netip.AddrPort
}

func NewRadioDaemon(control jsonapi.ControlURI, gnbs []jsonapi.ControlURI, reqPS []config.PDUSession, radio *Radio, pduSessions *PduSessions, psMan *PduSessionsManager, ueRanAddr netip.AddrPort) *RadioDaemon {
	return &RadioDaemon{
		Control:            control,
		Gnbs:               gnbs,
		ReqPS:              reqPS,
		Radio:              radio,
		PduSessions:        pduSessions,
		PduSessionsManager: psMan,
		UeRanAddr:          ueRanAddr,
	}
}

func (r *RadioDaemon) runDownlinkDaemon(ctx context.Context, srv *net.UDPConn, tun *water.Interface) error {
	if srv == nil {
		return fmt.Errorf("nil srv")
	}
	if tun == nil {
		return fmt.Errorf("nil tun iface")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, TUN_MTU)
			n, err := srv.Read(buf)
			if err != nil {
				return err
			}
			tun.Write(buf[:n])
		}
	}
	return nil
}

func (r *RadioDaemon) runUplinkDaemon(ctx context.Context, srv *net.UDPConn, tun *water.Interface) error {
	if srv == nil {
		return fmt.Errorf("nil srv")
	}
	if tun == nil {
		return fmt.Errorf("nil tun iface")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, TUN_MTU)
			n, err := tun.Read(buf)
			if err != nil {
				return err
			}
			r.PduSessionsManager.Write(buf[:n], srv)
		}
	}
	return nil
}

func (r *RadioDaemon) Start(ctx context.Context, tun *water.Interface) error {
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
	go func(ctx context.Context, srv *net.UDPConn, tun *water.Interface) {
		r.runDownlinkDaemon(ctx, srv, tun)
	}(ctx, srv, tun)
	go func(ctx context.Context, srv *net.UDPConn, tun *water.Interface) {
		r.runUplinkDaemon(ctx, srv, tun)
	}(ctx, srv, tun)

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
	for _, ps := range r.ReqPS {
		if err := r.PduSessions.InitEstablish(ctx, ps.Gnb, ps.Dnn); err != nil {
			return err
		}
	}
	return nil
}
