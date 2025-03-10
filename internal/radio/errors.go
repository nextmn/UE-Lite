// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package radio

import (
	"errors"
)

var (
	ErrNilTunIface             = errors.New("nil TUN interface")
	ErrNilUdpConn              = errors.New("nil UDP Connection")
	ErrUnknownGnb              = errors.New("Unknown gNB")
	ErrUnexpectedGnb           = errors.New("PDU session do not use the expected gNB")
	ErrPduSessionNotFound      = errors.New("No PDU Session found for this IP Address")
	ErrPduSessionAlreadyExists = errors.New("PDU session already exists")

	ErrUnsupportedPDUType = errors.New("Unsupported PDU Type")
	ErrMalformedPDU       = errors.New("Malformed PDU")
)
