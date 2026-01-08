// Copyright Louis Royer and the NextMN contributors. All rights reserved.
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

	go p.CreatePduSession(ps.Addr, ps.Header.Gnb)

	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})
}
