// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"net"
	"net/netip"
	"sync"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/sirupsen/logrus"
	"github.com/songgao/water/waterutil"
)

type PduSessionsManager struct {
	Links map[netip.Addr]jsonapi.ControlURI // UeIpAddr : Gnb control URI
	sync.Mutex
	isInit bool
	radio  *Radio
}

func NewPduSessionsManager(radio *Radio) *PduSessionsManager {
	return &PduSessionsManager{
		Links:  make(map[netip.Addr]jsonapi.ControlURI),
		isInit: false,
		radio:  radio,
	}
}

func (p *PduSessionsManager) Write(pkt []byte, srv *net.UDPConn) error {
	if !waterutil.IsIPv4(pkt) {
		return fmt.Errorf("not an IPv4 packet")
	}
	src, ok := netip.AddrFromSlice(waterutil.IPv4Source(pkt).To4())
	if !ok {
		return fmt.Errorf("error while retrieving ip addr")
	}
	gnb, ok := p.Links[src]
	if !ok {
		logrus.WithFields(
			logrus.Fields{
				"ip-addr": src,
			}).Trace("no pdu session found for this ip address")
		return fmt.Errorf("no pdu session found for this ip address")
	}
	ret := p.radio.Write(pkt, srv, gnb)
	if ret == nil {
		logrus.WithFields(
			logrus.Fields{
				"ip-addr": src,
			}).Trace("packet forwarded")
	}
	return ret

}

func (p *PduSessionsManager) DeletePduSession(ueIpAddr netip.Addr) error {
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
	}).Info("Removing link for PDU Session")
	p.Lock()
	defer p.Unlock()
	delete(p.Links, ueIpAddr)
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
	}).Debug("Creating ip address for new PDU Session")
	if err := runIP("addr", "del", fmt.Sprintf("%s/%d", ueIpAddr.String(), ueIpAddr.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ueIpAddr,
			"dev":        TUN_NAME,
		}).Error("Could not add ip address for new PDU Session")
		return err
	}
	return nil
}

func (p *PduSessionsManager) UpdatePduSession(ueIpAddr netip.Addr, newGnb jsonapi.ControlURI) {
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
		"new-gnb":    newGnb,
	}).Info("Updating link for PDU Session")
	p.Lock()
	defer p.Unlock()
	p.Links[ueIpAddr] = newGnb
}

func (p *PduSessionsManager) CreatePduSession(ueIpAddr netip.Addr, gnb jsonapi.ControlURI) error {
	p.Lock()
	defer p.Unlock()
	p.Links[ueIpAddr] = gnb

	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
	}).Debug("Creating ip address for new PDU Session")
	if err := runIP("addr", "add", fmt.Sprintf("%s/%d", ueIpAddr.String(), ueIpAddr.BitLen()), "dev", TUN_NAME); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"ue-ip-addr": ueIpAddr,
			"dev":        TUN_NAME,
		}).Error("Could not add ip address for new PDU Session")
		return err
	}
	return nil
}
