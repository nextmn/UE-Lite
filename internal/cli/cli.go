// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package cli

import (
	"github.com/nextmn/json-api/jsonapi"

	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"

	"github.com/gin-gonic/gin"
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

func (cli *Cli) Register(e *gin.Engine) {
	e.POST("/cli/radio/peer", cli.RadioPeer)
	e.POST("/cli/ps/establish", cli.PsEstablish)
}
