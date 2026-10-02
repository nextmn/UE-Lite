// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"encoding/json/v2"
	"net/http"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/sirupsen/logrus"
)

// Allow to peer to a gNB
func (r *Radio) Peer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	var peer n1n2.RadioPeerMsg
	if err := json.UnmarshalRead(req.Body, &peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		w.WriteHeader(http.StatusBadRequest)
		json.MarshalWrite(w, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	go r.HandlePeer(peer)
	w.WriteHeader(http.StatusAccepted)
	json.MarshalWrite(w, jsonapi.Message{Message: "please refer to logs for more information"})

}

func (r *Radio) HandlePeer(peer n1n2.RadioPeerMsg) {
	r.peerMap.Store(peer.Control.String(), peer.Data)
	logrus.WithFields(logrus.Fields{
		"peer-control": peer.Control.String(),
		"peer-ran":     peer.Data,
	}).Info("New peer radio link")
}
