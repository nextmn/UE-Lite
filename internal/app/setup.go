// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"time"

	"github.com/nextmn/ue-lite/internal/config"
	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"
	"github.com/nextmn/ue-lite/internal/tun"

	"github.com/sirupsen/logrus"
)

type Setup struct {
	config           *config.UEConfig
	httpServerEntity *HttpServerEntity
	radioDaemon      *radio.RadioDaemon
	ps               *session.PduSessions
	tunMan           *tun.TunManager
}

func NewSetup(config *config.UEConfig) *Setup {
	tunMan := tun.NewTunManager()
	r := radio.NewRadio(config.Control.Uri, tunMan, config.Ran.BindAddr, "go-github-nextmn-ue-lite")
	ps := session.NewPduSessions(config.Control.Uri, r, config.Ran.PDUSessions, "go-github-nextmn-ue-lite")
	return &Setup{
		config:           config,
		httpServerEntity: NewHttpServerEntity(config.Control.BindAddr, r, ps),
		radioDaemon:      radio.NewRadioDaemon(config.Control.Uri, config.Ran.Gnbs, r, config.Ran.BindAddr),
		ps:               ps,
		tunMan:           tunMan,
	}
}

func (s *Setup) Run(ctx context.Context) error {
	if err := s.httpServerEntity.Start(ctx); err != nil {
		return err
	}
	if err := s.tunMan.Start(ctx); err != nil {
		return err
	}
	logrus.Debug("TunMan started")
	if err := s.radioDaemon.Start(ctx); err != nil {
		return err
	}
	logrus.Debug("Radio Daemon started")
	if err := s.ps.Start(ctx); err != nil {
		return err
	}
	logrus.Debug("PsMan started")
	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	s.ps.WaitShutdown(ctxShutdown)
	s.radioDaemon.WaitShutdown(ctxShutdown)
	s.tunMan.WaitShutdown(ctxShutdown)
	s.httpServerEntity.WaitShutdown(ctxShutdown)
	return nil
}
