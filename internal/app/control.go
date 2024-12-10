// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/nextmn/ue-lite/internal/cli"
	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"

	"github.com/nextmn/json-api/healthcheck"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HttpServerEntity struct {
	srv   *http.Server
	ps    *session.PduSessions
	radio *radio.Radio
	cli   *cli.Cli
}

func NewHttpServerEntity(bindAddr netip.AddrPort, r *radio.Radio, ps *session.PduSessions) *HttpServerEntity {
	c := cli.NewCli(r, ps)
	// TODO: gin.SetMode(gin.DebugMode) / gin.SetMode(gin.ReleaseMode) depending on log level
	h := gin.Default()
	h.GET("/status", Status)

	// CLI
	h.POST("/cli/radio/peer", c.RadioPeer)
	h.POST("/cli/ps/establish", c.PsEstablish)

	// Radio
	h.POST("/radio/peer", r.Peer)

	// Pdu Session
	h.POST("/ps/establishment-accept", ps.EstablishmentAccept)

	logrus.WithFields(logrus.Fields{"http-addr": bindAddr}).Info("HTTP Server created")
	e := HttpServerEntity{
		srv: &http.Server{
			Addr:    bindAddr.String(),
			Handler: h,
		},
		ps:    ps,
		radio: r,
		cli:   c,
	}
	return &e
}

func (e *HttpServerEntity) Start() error {
	l, err := net.Listen("tcp", e.srv.Addr)
	if err != nil {
		return err
	}
	go func(ln net.Listener) {
		logrus.Info("Starting HTTP Server")
		if err := e.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("Http Server error")
		}
	}(l)
	return nil
}

func (e *HttpServerEntity) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // context.Background() is already Done()
	defer cancel()
	if err := e.srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Info("HTTP Server Shutdown")
	}
}

// get status of the controller
func Status(c *gin.Context) {
	status := healthcheck.Status{
		Ready: true,
	}
	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, status)
}
