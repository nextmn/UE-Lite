// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PduSessions) Status(c *gin.Context) {
	sessions := p.radio.GetRoutes()

	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, sessions)
}
