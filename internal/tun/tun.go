// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tun

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

const (
	TUN_NAME = "nextmn-ue-lite"
	TUN_MTU  = 1400
)

type TunManager struct {
	ready  bool
	name   string
	tun    *water.Interface
	closed chan struct{}
	used   sync.WaitGroup
}

func NewTunManager() *TunManager {
	return &TunManager{
		closed: make(chan struct{}),
	}
}

// Get a tun interface.
// Don't forget to run CloseTun when no longer in use
func (t *TunManager) OpenTun() *water.Interface {
	t.used.Add(1)
	return t.tun
}

func (t *TunManager) CloseTun() {
	t.used.Done()
}

func (t *TunManager) Start(ctx context.Context) error {
	tun, err := newTunIface(ctx)
	t.tun = tun
	if err != nil {
		return err
	}
	t.name = t.tun.Name()
	t.ready = true
	go func(ctx context.Context) {
		<-ctx.Done()
		t.used.Wait() // Do not delete tun iface until all tuns are closed

		ctxDel := context.WithoutCancel(ctx) // required to force cleanup
		if err := runIPTables(ctxDel, "-D", "OUTPUT", "-o", t.name, "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP"); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{"interface": t.name}).Error("Error while removing iptables rules")
		}
		t.ready = false
		close(t.closed)
	}(ctx)
	return err
}

func (t *TunManager) WaitShutdown(ctx context.Context) error {
	if !t.ready {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.closed:
		return nil
	}
}

func newTunIface(ctx context.Context) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = TUN_NAME
	iface, err := water.New(config)
	if err != nil {
		logrus.WithError(err).Error("Unable to allocate TUN interface")
		return nil, err
	}
	if err := runIP(ctx, "link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(TUN_MTU)); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"mtu":       TUN_MTU,
			"interface": iface.Name(),
		}).Error("Unable to set MTU")
		return nil, err
	}
	if err := runIP(ctx, "link", "set", "dev", iface.Name(), "up"); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"interface": iface.Name(),
		}).Error("Unable to set interface up")
		return nil, err
	}
	// TODO: add proto "nextmn-lite-ue"
	if err := runIP(ctx, "route", "replace", "default", "dev", iface.Name()); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"interface": iface.Name(),
		}).Error("Unable to set default route")
		return nil, err
	}
	if err := runIPTables(ctx, "-A", "OUTPUT", "-o", iface.Name(), "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP"); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"interface": iface.Name()}).Error("Error while setting iptable rule to drop icmp redirects")
		return nil, err
	}
	return iface, nil
}

func (t *TunManager) DelIp(ctx context.Context, ip netip.Addr) error {
	if err := runIP(ctx, "addr", "del", fmt.Sprintf("%s/%d", ip.String(), ip.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ip,
			"dev":        TUN_NAME,
		}).Error("Could not remove ip address")
		return err
	}
	return nil
}
func (t *TunManager) AddIp(ctx context.Context, ip netip.Addr) error {
	if err := runIP(ctx, "addr", "add", fmt.Sprintf("%s/%d", ip.String(), ip.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ip,
			"dev":        TUN_NAME,
		}).Error("Could not add ip address for new PDU Session")
		return err
	}
	return nil
}
