// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"

	"github.com/nextmn/ue-lite/internal/config"
)

type Setup struct {
	config           *config.UEConfig
	httpServerEntity *HttpServerEntity
	radioDaemon      *RadioDaemon
	psMan            *PduSessionsManager
}

func NewSetup(config *config.UEConfig) *Setup {
	radio := NewRadio(config.Control.Uri, config.Ran.BindAddr, "go-github-nextmn-ue-lite")
	psMan := NewPduSessionsManager(radio)
	ps := NewPduSessions(config.Control.Uri, psMan, "go-github-nextmn-ue-lite")
	return &Setup{
		config:           config,
		httpServerEntity: NewHttpServerEntity(config.Control.BindAddr, radio, ps),
		radioDaemon:      NewRadioDaemon(config.Control.Uri, config.Ran.Gnbs, config.Ran.PDUSessions, radio, ps, psMan, config.Ran.BindAddr),
		psMan:            psMan,
	}
}

func (s *Setup) Init(ctx context.Context) error {
	if err := s.httpServerEntity.Start(); err != nil {
		return err
	}
	tun, err := NewTunIface()
	if err != nil {
		return err
	}
	if err := s.radioDaemon.Start(ctx, tun); err != nil {
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
