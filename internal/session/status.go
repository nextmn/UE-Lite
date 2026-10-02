// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package session

import (
	"encoding/json/v2"
	"net/http"
)

func (p *PduSessions) Status(w http.ResponseWriter, req *http.Request) {
	sessions := p.radio.GetRoutes()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.MarshalWrite(w, sessions)
}
