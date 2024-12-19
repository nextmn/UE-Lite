// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
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
		"gnb-source": ps.GNBSource.String(),
		"gnb-target": ps.GNBTarget.String(),
	}).Info("New Handover Command")

	go p.HandleHandoverCommand(ps)

	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})
}

func (p *PduSessions) HandleHandoverCommand(m n1n2.HandoverCommand) {
	if m.GNBSource == m.GNBTarget {
		logrus.WithFields(logrus.Fields{
			"gnb": m.GNBSource.String(),
		}).Error("Handover Command: source and target gNBs are not different.")
		// TODO: notify gNB/CP of failure?
		return
	}

	// TODO: update pdu sessions atomically
	for _, session := range m.Sessions {
		if err := p.UpdatePduSession(session, m.GNBSource, m.GNBTarget); err != nil {
			// TODO: notify gNB/CP of failure?
			continue
		}
	}
}
