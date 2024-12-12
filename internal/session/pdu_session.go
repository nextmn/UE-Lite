// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/nextmn/ue-lite/internal/config"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PduSessions struct {
	Control   jsonapi.ControlURI
	Client    http.Client
	UserAgent string
	psMan     *PduSessionsManager
	reqPs     []config.PDUSession
}

func NewPduSessions(control jsonapi.ControlURI, psMan *PduSessionsManager, reqPs []config.PDUSession, userAgent string) *PduSessions {
	return &PduSessions{
		Client:    http.Client{},
		Control:   control,
		UserAgent: userAgent,
		psMan:     psMan,
		reqPs:     reqPs,
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

	logrus.WithFields(logrus.Fields{
		"gnb":     ps.Header.Gnb.String(),
		"ip-addr": ps.Addr,
	}).Info("New PDU Session")

	p.psMan.CreatePduSession(ps.Addr, ps.Header.Gnb)

	c.Status(http.StatusNoContent)
}

func (p *PduSessions) Start(ctx context.Context) error {
	for _, ps := range p.reqPs {
		if err := p.InitEstablish(ctx, ps.Gnb, ps.Dnn); err != nil {
			return err
		}
	}
	return nil
}

func (p *PduSessions) WaitShutdown(ctx context.Context) error {
	// nothing to do
	return nil
}
