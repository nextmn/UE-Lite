// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package cli

import (
	"net/http"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
