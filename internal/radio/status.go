// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"net/http"
	"net/netip"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (r *Radio) Status(c *gin.Context) {
	peers := make(map[string]netip.AddrPort)
	r.peerMap.Range(func(key, value any) bool {
		peers[key.(string)] = value.(netip.AddrPort)
		logrus.WithFields(logrus.Fields{
			"key":   key.(string),
			"value": value.(netip.AddrPort),
		}).Trace("Creating radio/status response")
		return true
	})

	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, peers)
}
