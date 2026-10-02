// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package cli

import (
	"net/http"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"
)

type Cli struct {
	Radio       *radio.Radio
	PduSessions *session.PduSessions
}

type CliPeerMsg struct {
	Gnb jsonapi.ControlURI `json:"gnb"`
	Dnn string             `json:"dnn"`
}

func (cli Cli) Handler() http.Handler {
	sm := http.NewServeMux()
	sm.HandleFunc("POST /radio/peer", cli.RadioPeer)
	sm.HandleFunc("POST /ps/establish", cli.PsEstablish)
	return sm
}
