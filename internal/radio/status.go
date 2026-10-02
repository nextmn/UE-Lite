// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"encoding/json/v2"
	"net/http"
	"net/netip"

	"github.com/sirupsen/logrus"
)

func (r *Radio) Status(w http.ResponseWriter, req *http.Request) {
	peers := make(map[string]netip.AddrPort)
	r.peerMap.Range(func(key, value any) bool {
		peers[key.(string)] = value.(netip.AddrPort)
		logrus.WithFields(logrus.Fields{
			"key":   key.(string),
			"value": value.(netip.AddrPort),
		}).Trace("Creating radio/status response")
		return true
	})

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.MarshalWrite(w, peers)
}
