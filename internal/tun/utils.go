// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tun

import (
	"context"
	"fmt"
	"os/exec"
)

// Run ip command
func runIP(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ip", args...)
	cmd.Env = []string{}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %s: %w", cmd.Args, err)
	}
	return nil
}

// Run iptables command
func runIPTables(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "iptables", args...)
	cmd.Env = []string{}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running %s: %w", cmd.Args, err)
	}
	return nil
}
