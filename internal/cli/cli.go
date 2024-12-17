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
	go cli.HandleRadioPeer(peer)
	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})
}

func (cli *Cli) HandleRadioPeer(peer CliPeerMsg) {
	if err := cli.Radio.InitPeer(peer.Gnb); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gnb": peer.Gnb,
			"dnn": peer.Dnn,
		}).Error("Could not perform Radio Peer Init")
		return
	}
	// TODO: handle gnb failure
}

func (cli *Cli) PsEstablish(c *gin.Context) {
	var peer CliPeerMsg
	if err := c.BindJSON(&peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	go cli.HandlePsEstablish(peer)
	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})
}

func (cli *Cli) HandlePsEstablish(peer CliPeerMsg) {
	// TODO: first, check if radio link is established
	if err := cli.PduSessions.InitEstablish(peer.Gnb, peer.Dnn); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"gnb": peer.Gnb,
			"dnn": peer.Dnn,
		}).Error("Could not perform PDU Session Establishment")
		return
	}
	// TODO: handle gnb failure

}
