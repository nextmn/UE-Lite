// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"encoding/json"
	"net/http"
	"net/netip"

	"github.com/nextmn/json-api/jsonapi"

	"github.com/gin-gonic/gin"
)

func (r *Radio) Status(c *gin.Context) {
	peers := make(map[jsonapi.ControlURI]netip.Addr)
	r.peerMap.Range(func(key, value any) bool {
		peers[key.(jsonapi.ControlURI)] = value.(netip.Addr)
		return true
	})
	j, err := json.Marshal(peers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsonapi.MessageWithError{Message: "could not marshal peers map", Error: err})
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, j)
}
