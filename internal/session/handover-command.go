// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (p *PduSessions) HandoverCommand(c *gin.Context) {
	var ps n1n2.HandoverCommand
	if err := c.BindJSON(&ps); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}

	logrus.WithFields(logrus.Fields{
		"gnb-source": ps.SourceGnb.String(),
		"gnb-target": ps.TargetGnb.String(),
	}).Info("New Handover Command")

	go p.HandleHandoverCommand(ps)

	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})
}

func (p *PduSessions) HandleHandoverCommand(m n1n2.HandoverCommand) {
	ctx := p.Context()
	if m.SourceGnb == m.TargetGnb {
		logrus.WithFields(logrus.Fields{
			"gnb": m.SourceGnb.String(),
		}).Error("Handover Command: source and target gNBs are not different.")
		// TODO: notify gNB/CP of failure?
		return
	}

	sessions := make([]n1n2.Session, len(m.Sessions))
	// TODO: update pdu sessions atomically
	for i, session := range m.Sessions {
		if err := p.UpdatePduSession(session.Addr, m.SourceGnb, m.TargetGnb); err != nil {
			// TODO: notify gNB/CP of failure?
			logrus.WithError(err).WithFields(logrus.Fields{
				"ue-addr":    session.Addr,
				"source-gnb": m.SourceGnb.String(),
				"target-gnb": m.TargetGnb.String(),
			}).Error("Handover failure")
			continue
		}
		sessions[i] = n1n2.Session{
			Addr: session.Addr,
			Dnn:  session.Dnn,
		}
	}

	// Send Handover Confirm
	resp := n1n2.HandoverConfirm{
		// Header
		UeCtrl: m.UeCtrl,
		Cp:     m.Cp,

		// Payload
		Sessions:  sessions,
		SourceGnb: m.SourceGnb,
		TargetGnb: m.TargetGnb,
	}
	reqBody, err := json.Marshal(resp)
	if err != nil {
		logrus.WithError(err).Error("Could not marshal n1n2.HandoverConfirm")
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.TargetGnb.JoinPath("ps/handover-confirm").String(), bytes.NewBuffer(reqBody))
	if err != nil {
		logrus.WithError(err).Error("Could not create request for ps/handover-confirm")
		return
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	if _, err := p.Client.Do(req); err != nil {
		logrus.WithError(err).Error("Could not send ps/handover-confirm")
	}
}
