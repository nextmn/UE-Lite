// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"

	"github.com/nextmn/ue-lite/internal/config"
	"github.com/nextmn/ue-lite/internal/radio"
	"github.com/nextmn/ue-lite/internal/session"
	"github.com/nextmn/ue-lite/internal/tun"
)

type Setup struct {
	config           *config.UEConfig
	httpServerEntity *HttpServerEntity
	radioDaemon      *radio.RadioDaemon
	ps               *session.PduSessions
	tunMan           *tun.TunManager
}

func NewSetup(config *config.UEConfig) *Setup {
	r := radio.NewRadio(config.Control.Uri, config.Ran.BindAddr, "go-github-nextmn-ue-lite")
	tunMan := tun.NewTunManager()
	psMan := session.NewPduSessionsManager(tunMan)
	ps := session.NewPduSessions(config.Control.Uri, psMan, config.Ran.PDUSessions, "go-github-nextmn-ue-lite")
	return &Setup{
		config:           config,
		httpServerEntity: NewHttpServerEntity(config.Control.BindAddr, r, ps),
		radioDaemon:      radio.NewRadioDaemon(config.Control.Uri, config.Ran.Gnbs, r, psMan, tunMan, config.Ran.BindAddr),
		ps:               ps,
		tunMan:           tunMan,
	}
}

func (s *Setup) Init(ctx context.Context) error {
	if err := s.httpServerEntity.Start(); err != nil {
		return err
	}
	if err := s.tunMan.Start(ctx); err != nil {
		return err
	}
	if err := s.radioDaemon.Start(ctx); err != nil {
		return err
	}
	if err := s.ps.Start(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Setup) Run(ctx context.Context) error {
	defer s.Exit()
	if err := s.Init(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	}
}

func (s *Setup) Exit() error {
	s.httpServerEntity.Stop()
	return nil
}
