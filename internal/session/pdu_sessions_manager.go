// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"net/netip"
	"sync"

	"github.com/nextmn/ue-lite/internal/tun"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/sirupsen/logrus"
)

type PduSessionsManager struct {
	Links map[netip.Addr]jsonapi.ControlURI // UeIpAddr : Gnb control URI
	sync.Mutex
	isInit bool
	tun    *tun.TunManager
}

func NewPduSessionsManager(tunMan *tun.TunManager) *PduSessionsManager {
	return &PduSessionsManager{
		Links:  make(map[netip.Addr]jsonapi.ControlURI),
		isInit: false,
		tun:    tunMan,
	}
}

func (p *PduSessionsManager) LinkedGnb(src netip.Addr) (jsonapi.ControlURI, error) {
	gnb, ok := p.Links[src]
	if !ok {
		logrus.WithFields(
			logrus.Fields{
				"ip-addr": src,
			}).Trace("no pdu session found for this ip address")
		return jsonapi.ControlURI{}, ErrPduSessionNotFound
	}
	return gnb, nil

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
	if err := p.tun.DelIp(ueIpAddr); err != nil {
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
	if err := p.tun.AddIp(ueIpAddr); err != nil {
		return err
	}
	return nil
}
