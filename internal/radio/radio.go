// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
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

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Radio struct {
	Client    http.Client
	peerMap   sync.Map // key: gnb control uri ; value: gnb ran ip address
	Control   jsonapi.ControlURI
	Data      netip.AddrPort
	UserAgent string

	// not exported because must not be modified
	ctx context.Context
}

func NewRadio(control jsonapi.ControlURI, data netip.AddrPort, userAgent string) *Radio {
	return &Radio{
		peerMap:   sync.Map{},
		Client:    http.Client{},
		Control:   control,
		Data:      data,
		UserAgent: userAgent,
	}
}

func (r *Radio) Write(pkt []byte, srv *net.UDPConn, gnb jsonapi.ControlURI) error {
	gnbRan, ok := r.peerMap.Load(gnb)
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

func (r *Radio) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}
func (r *Radio) Init(ctx context.Context) error {
	if ctx == nil {
		return ErrNilCtx
	}
	r.ctx = ctx
	return nil
}

func (r *Radio) Register(e *gin.Engine) {
	e.GET("/radio", r.Status)
	e.POST("/radio/peer", r.Peer)
}
