// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PduSessions struct {
	PduSessionsMap     sync.Map // key: overlay ip address, value: gnb control uri
	Control            jsonapi.ControlURI
	Client             http.Client
	UserAgent          string
	PduSessionsManager *PduSessionsManager
}

func NewPduSessions(control jsonapi.ControlURI, pduSessionsManager *PduSessionsManager, userAgent string) *PduSessions {
	return &PduSessions{
		Client:             http.Client{},
		PduSessionsMap:     sync.Map{},
		Control:            control,
		UserAgent:          userAgent,
		PduSessionsManager: pduSessionsManager,
	}
}

func (p *PduSessions) InitEstablish(ctx context.Context, gnb jsonapi.ControlURI, dnn string) error {
	logrus.WithFields(logrus.Fields{
		"gnb": gnb.String(),
	}).Info("Creating new PDU Session")

	msg := n1n2.PduSessionEstabReqMsg{
		Ue:  p.Control,
		Gnb: gnb,
		Dnn: dnn,
	}
	reqBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gnb.JoinPath("ps/establishment-request").String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// get status of the controller
func (p *PduSessions) EstablishmentAccept(c *gin.Context) {
	var ps n1n2.PduSessionEstabAcceptMsg
	if err := c.BindJSON(&ps); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	p.PduSessionsMap.Store(ps.Addr, ps.Header.Gnb)

	logrus.WithFields(logrus.Fields{
		"gnb":     ps.Header.Gnb.String(),
		"ip-addr": ps.Addr,
	}).Info("New PDU Session")

	p.PduSessionsManager.CreatePduSession(ps.Addr, ps.Header.Gnb)

	c.Status(http.StatusNoContent)
}
