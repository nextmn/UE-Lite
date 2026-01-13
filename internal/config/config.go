// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"

	"github.com/nextmn/json-api/jsonapi"

	"gopkg.in/yaml.v3"
)

var ErrEmptyConfigFilepath = errors.New("`$CONFIG` is not set, `config` flag is not set, and default config file does not exist")

func ParseConf(file string) (*UEConfig, error) {
	var conf UEConfig
	if v, ok := os.LookupEnv("CONFIG"); ok {
		err := yaml.Unmarshal([]byte(v), &conf)
		if err != nil {
			return nil, err
		}
		return &conf, nil
	}
	if file == "" {
		return nil, ErrEmptyConfigFilepath
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

type UEConfig struct {
	Control Control `yaml:"control"`
	Ran     Ran     `yaml:"ran"`
	Logger  *Logger `yaml:"logger,omitempty"`
}

type Control struct {
	Uri      jsonapi.ControlURI `yaml:"uri"`       // may contain domain name instead of ip address
	BindAddr netip.AddrPort     `yaml:"bind-addr"` // in the form `ip:port`
}

type Ran struct {
	BindAddr    netip.AddrPort       `yaml:"bind-addr"`    // in the form ip:port
	Gnbs        []jsonapi.ControlURI `yaml:"gnbs"`         // list of gnb used
	PDUSessions []PDUSession         `yaml:"pdu-sessions"` // list of pdu sessions that will be established
}

type PDUSession struct {
	Gnb jsonapi.ControlURI `yaml:"gnb"`
	Dnn string             `yaml:"dnn"`
}
