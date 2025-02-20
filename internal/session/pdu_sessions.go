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
	"net/netip"

	"github.com/nextmn/ue-lite/internal/common"
	"github.com/nextmn/ue-lite/internal/config"
	"github.com/nextmn/ue-lite/internal/radio"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PduSessions struct {
	common.WithContext

	Control   jsonapi.ControlURI
	Client    http.Client
	UserAgent string
	reqPs     []config.PDUSession
	radio     *radio.Radio
}

func NewPduSessions(control jsonapi.ControlURI, r *radio.Radio, reqPs []config.PDUSession, userAgent string) *PduSessions {
	return &PduSessions{
		Client:    http.Client{},
		Control:   control,
		UserAgent: userAgent,
		reqPs:     reqPs,
		radio:     r,
	}
}

func (p *PduSessions) Register(e *gin.Engine) {
	e.GET("/ps", p.Status)
	e.POST("/ps/establishment-accept", p.EstablishmentAccept)
	e.POST("/ps/handover-command", p.HandoverCommand)
}

func (p *PduSessions) InitEstablish(gnb jsonapi.ControlURI, dnn string) error {
	ctx := p.Context()
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

func (p *PduSessions) Start(ctx context.Context) error {
	if err := p.InitContext(ctx); err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"number-of-pdu-sessions-requested": len(p.reqPs),
	}).Info("Starting PDU Sessions Manager")
	for _, ps := range p.reqPs {
		if err := p.InitEstablish(ps.Gnb, ps.Dnn); err != nil {
			return err
		}
	}
	return nil
}

func (p *PduSessions) WaitShutdown(ctx context.Context) error {
	// nothing to do
	return nil
}

func (p *PduSessions) DeletePduSession(ueIpAddr netip.Addr) error {
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
	}).Debug("Removing PDU Session")
	return p.radio.DelRoute(ueIpAddr)
}

func (p *PduSessions) UpdatePduSession(ueIpAddr netip.Addr, oldGnb jsonapi.ControlURI, newGnb jsonapi.ControlURI) error {
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
		"old-gnb":    oldGnb.String(),
		"new-gnb":    newGnb.String(),
	}).Info("Updating PDU Session")
	return p.radio.UpdateRoute(ueIpAddr, oldGnb, newGnb)
}

func (p *PduSessions) CreatePduSession(ueIpAddr netip.Addr, gnb jsonapi.ControlURI) error {
	logrus.WithFields(logrus.Fields{
		"ue-ip-addr": ueIpAddr,
	}).Debug("Creating new PDU Session")
	return p.radio.AddRoute(ueIpAddr, gnb)
}
