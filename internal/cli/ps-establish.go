// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package cli

import (
	"encoding/json/v2"
	"net/http"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/sirupsen/logrus"
)

func (cli Cli) PsEstablish(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	var peer CliPeerMsg
	if err := json.UnmarshalRead(req.Body, &peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		w.WriteHeader(http.StatusBadRequest)
		json.MarshalWrite(w, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	go cli.HandlePsEstablish(peer)
	w.WriteHeader(http.StatusAccepted)
	json.MarshalWrite(w, jsonapi.Message{Message: "please refer to logs for more information"})
}

func (cli Cli) HandlePsEstablish(peer CliPeerMsg) {
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
