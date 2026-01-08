// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/netip"
	"sync"

	"github.com/nextmn/ue-lite/internal/common"
	"github.com/nextmn/ue-lite/internal/tun"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Radio struct {
	common.WithContext

	Client       http.Client
	peerMap      sync.Map // key: gnb control uri (string); value: gnb ran ip address
	routingTable sync.Map // key: ueIp; value gnb control uri
	Tun          *tun.TunManager
	Control      jsonapi.ControlURI
	Data         netip.AddrPort
	UserAgent    string

	// not exported because must not be modified
	ctx context.Context
}

func NewRadio(control jsonapi.ControlURI, tunMan *tun.TunManager, data netip.AddrPort, userAgent string) *Radio {
	return &Radio{
		peerMap:      sync.Map{},
		routingTable: sync.Map{},
		Client:       http.Client{},
		Control:      control,
		Data:         data,
		UserAgent:    userAgent,
		Tun:          tunMan,
	}
}

// AddRoute creates a route to the gNB for this PDU session, including configuration of iproute2 interface
func (r *Radio) AddRoute(ueIp netip.Addr, gnb jsonapi.ControlURI) error {
	if _, ok := r.peerMap.Load(gnb.String()); !ok {
		return ErrUnknownGnb
	}
	if _, loaded := r.routingTable.LoadOrStore(ueIp, gnb); loaded {
		return ErrPduSessionAlreadyExists
	}
	return r.Tun.AddIp(ueIp)
}

// DelRoute remove the route to the gNB for this PDU session, including (de-)configuration of iproute2 interface
func (r *Radio) DelRoute(ueIp netip.Addr) error {
	r.routingTable.Delete(ueIp)
	return r.Tun.DelIp(ueIp)
}

// UpdateRoute updates the route to the gNB for this PDU Session
func (r *Radio) UpdateRoute(ueIp netip.Addr, oldGnb jsonapi.ControlURI, newGnb jsonapi.ControlURI) error {
	if _, ok := r.peerMap.Load(newGnb.String()); !ok {
		return ErrUnknownGnb
	}
	old, ok := r.routingTable.Load(ueIp)
	if !ok {
		return ErrPduSessionNotFound
	}
	oldT := old.(jsonapi.ControlURI)
	if oldT.String() != oldGnb.String() {
		return ErrUnexpectedGnb
	}

	r.routingTable.Store(ueIp, newGnb)
	return nil
}

func (r *Radio) GetRoutes() map[netip.Addr]jsonapi.ControlURI {
	sessions := make(map[netip.Addr]jsonapi.ControlURI)
	r.routingTable.Range(func(key, value any) bool {
		sessions[key.(netip.Addr)] = value.(jsonapi.ControlURI)
		logrus.WithFields(logrus.Fields{
			"key":   key.(netip.Addr),
			"value": value.(jsonapi.ControlURI),
		}).Trace("Creating ps/status response")
		return true
	})
	return sessions
}

func (r *Radio) Write(pkt []byte, srv *net.UDPConn, ue netip.Addr) error {
	gnb, ok := r.routingTable.Load(ue)
	if !ok {
		logrus.Trace("PDU Session not found for this IP Address")
		return ErrPduSessionNotFound
	}
	gnbT := gnb.(jsonapi.ControlURI)
	gnbRan, ok := r.peerMap.Load(gnbT.String())
	if !ok {
		logrus.Trace("Unknown gnb")
		return ErrUnknownGnb
	}

	_, err := srv.WriteToUDPAddrPort(pkt, gnbRan.(netip.AddrPort))

	return err
}

func (r *Radio) InitPeer(gnb jsonapi.ControlURI) error {
	ctx := r.Context()
	logrus.WithFields(logrus.Fields{
		"gnb": gnb.String(),
	}).Info("Creating radio link with a new gNB")

	msg := n1n2.RadioPeerMsg{
		Control: r.Control,
		Data:    r.Data,
	}

	reqBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gnb.JoinPath("radio/peer").String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (r *Radio) Register(e *gin.Engine) {
	e.GET("/radio", r.Status)
	e.POST("/radio/peer", r.Peer)
}
