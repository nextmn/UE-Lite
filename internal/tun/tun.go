// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tun

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

const (
	TUN_NAME = "nextmn-ue-lite"
	TUN_MTU  = 1400
)

type TunManager struct {
	ready bool
	name  string
	Tun   *water.Interface
}

func NewTunManager() *TunManager {
	return &TunManager{}
}

func (t *TunManager) Start(ctx context.Context) error {
	tun, err := newTunIface()
	t.Tun = tun
	if err != nil {
		return err
	}
	t.ready = true
	t.name = t.Tun.Name()
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			err = runIPTables("-D", "OUTPUT", "-o", t.name, "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP")
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{"interface": t.name}).Error("Error while removing iptables rules")
				t.ready = false
			}
		}
	}(ctx)
	return err
}

func newTunIface() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = TUN_NAME
	iface, err := water.New(config)
	if nil != err {
		logrus.WithError(err).Error("Unable to allocate TUN interface")
		return nil, err
	}
	err = runIP("link", "set", "dev", iface.Name(), "mtu", strconv.Itoa(TUN_MTU))
	if nil != err {
		logrus.WithError(err).WithFields(logrus.Fields{
			"mtu":       TUN_MTU,
			"interface": iface.Name(),
		}).Error("Unable to set MTU")
		return nil, err
	}
	err = runIP("link", "set", "dev", iface.Name(), "up")
	if nil != err {
		logrus.WithError(err).WithFields(logrus.Fields{
			"interface": iface.Name(),
		}).Error("Unable to set interface up")
		return nil, err
	}
	// TODO: add proto "nextmn-lite-ue"
	err = runIP("route", "replace", "default", "dev", iface.Name())
	if nil != err {
		logrus.WithError(err).WithFields(logrus.Fields{
			"interface": iface.Name(),
		}).Error("Unable to set default route")
		return nil, err
	}
	err = runIPTables("-A", "OUTPUT", "-o", iface.Name(), "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP")
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"interface": iface.Name()}).Error("Error while setting iptable rule to drop icmp redirects")
		return nil, err
	}
	return iface, nil
}

func (t *TunManager) DelIp(ip netip.Addr) error {
	if err := runIP("addr", "del", fmt.Sprintf("%s/%d", ip.String(), ip.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ip,
			"dev":        TUN_NAME,
		}).Error("Could not remove ip address")
		return err
	}
	return nil
}
func (t *TunManager) AddIp(ip netip.Addr) error {
	if err := runIP("addr", "add", fmt.Sprintf("%s/%d", ip.String(), ip.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ip,
			"dev":        TUN_NAME,
		}).Error("Could not add ip address for new PDU Session")
		return err
	}
	return nil
}
