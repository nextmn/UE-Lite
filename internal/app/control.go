// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"encoding/json/v2"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/nextmn/ue-lite/internal/cli"
	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"

	"github.com/nextmn/json-api/healthcheck"

	"github.com/sirupsen/logrus"
)

type HttpServerEntity struct {
	srv    *http.Server
	closed chan struct{}
}

func NewHttpServerEntity(bindAddr netip.AddrPort, r *radio.Radio, ps *session.PduSessions) *HttpServerEntity {
	c := cli.Cli{Radio: r, PduSessions: ps}
	h := http.NewServeMux()
	h.HandleFunc("GET /status", Status)
	h.Handle("/cli", http.StripPrefix("/cli", c.Handler()))
	h.Handle("/radio", http.StripPrefix("/radio", r.Handler()))
	h.Handle("/ps", http.StripPrefix("/ps", r.Handler()))

	logrus.WithFields(logrus.Fields{"http-addr": bindAddr}).Info("HTTP Server created")
	e := HttpServerEntity{
		srv: &http.Server{
			Addr:    bindAddr.String(),
			Handler: h,
		},
		closed: make(chan struct{}),
	}
	return &e
}

func (e *HttpServerEntity) Start(ctx context.Context) error {
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
	go func(ctx context.Context) {
		defer close(e.closed)
		<-ctx.Done()
		ctxTimeout, cancel := context.WithTimeout(context.WithoutCancel(ctx), 100*time.Millisecond)
		defer cancel()
		if err := e.srv.Shutdown(ctxTimeout); err == nil {
			logrus.Info("HTTP Server Shutdown")
		}
	}(ctx)
	return nil
}

func (e *HttpServerEntity) WaitShutdown(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.closed:
		return nil
	}
}

// get status of the controller
func Status(w http.ResponseWriter, req *http.Request) {
	status := healthcheck.Status{
		Ready: true,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.MarshalWrite(w, status)
}
