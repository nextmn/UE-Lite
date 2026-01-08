// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"net/http"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n1n2"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Allow to peer to a gNB
func (r *Radio) Peer(c *gin.Context) {
	var peer n1n2.RadioPeerMsg
	if err := c.BindJSON(&peer); err != nil {
		logrus.WithError(err).Error("could not deserialize")
		c.JSON(http.StatusBadRequest, jsonapi.MessageWithError{Message: "could not deserialize", Error: err})
		return
	}
	go r.HandlePeer(peer)
	c.JSON(http.StatusAccepted, jsonapi.Message{Message: "please refer to logs for more information"})

}

func (r *Radio) HandlePeer(peer n1n2.RadioPeerMsg) {
	r.peerMap.Store(peer.Control.String(), peer.Data)
	logrus.WithFields(logrus.Fields{
		"peer-control": peer.Control.String(),
		"peer-ran":     peer.Data,
	}).Info("New peer radio link")
}
