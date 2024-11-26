// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("Unknown gnb")
	}

	_, err := srv.WriteToUDPAddrPort(pkt, gnbRan.(netip.AddrPort))

	return err
}

func (r *Radio) InitPeer(ctx context.Context, gnb jsonapi.ControlURI) error {
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

// Allow to peer to a gNB
func (r *Radio) Peer(c *gin.Context) {
	var peer n1n2.RadioPeerMsg
	if err := c.BindJSON(&peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	r.peerMap.Store(peer.Control, peer.Data)
	logrus.WithFields(logrus.Fields{
		"peer-control": peer.Control.String(),
		"peer-ran":     peer.Data,
	}).Info("New peer radio link")

	c.Status(http.StatusNoContent)
}
