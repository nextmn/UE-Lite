// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"encoding/json/v2"
	"net/http"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/sirupsen/logrus"
)

// get status of the controller
func (p *PduSessions) EstablishmentAccept(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	var ps n1n2.PduSessionEstabAcceptMsg
	if err := json.UnmarshalRead(req.Body, &ps); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		w.WriteHeader(http.StatusBadRequest)
		json.MarshalWrite(w, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}

	logrus.WithFields(logrus.Fields{
		"gnb":     ps.Header.Gnb.String(),
		"ip-addr": ps.Addr,
	}).Info("New PDU Session")

	go p.CreatePduSession(ps.Addr, ps.Header.Gnb)

	w.WriteHeader(http.StatusAccepted)
	json.MarshalWrite(w, jsonapi.Message{Message: "please refer to logs for more information"})
}
