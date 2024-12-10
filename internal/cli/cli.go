// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package cli

import (
	"net/http"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Cli struct {
	Radio       *radio.Radio
	PduSessions *session.PduSessions
}

func NewCli(radio *radio.Radio, pduSessions *session.PduSessions) *Cli {
	return &Cli{
		Radio:       radio,
		PduSessions: pduSessions,
	}
}

type CliPeerMsg struct {
	Gnb jsonapi.ControlURI `json:"gnb"`
	Dnn string             `json:"dnn"`
}

// Allow to peer to a gNB
func (cli *Cli) RadioPeer(c *gin.Context) {
	var peer CliPeerMsg
	if err := c.BindJSON(&peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	if err := cli.Radio.InitPeer(c, peer.Gnb); err != nil {
		logrus.WithError(err).Error("could not perform radio peer init")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not perform radio peer init", Error: err})
		return
	}

	// TODO: handle gnb failure

	c.Status(http.StatusNoContent)
}

func (cli *Cli) PsEstablish(c *gin.Context) {
	var peer CliPeerMsg
	if err := c.BindJSON(&peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	// TODO: first, check if radio link is established

	if err := cli.PduSessions.InitEstablish(c, peer.Gnb, peer.Dnn); err != nil {
		logrus.WithError(err).Error("could not perform pdu session establishment")
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not perform pdu session establishment", Error: err})
		return
	}

	// TODO: handle gnb failure

	c.Status(http.StatusNoContent)
}
