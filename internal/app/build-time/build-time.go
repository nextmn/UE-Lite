// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package buildtime

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/urfave/cli/v3"
)

// Get the build time if defined, or unix epoch
func getBuildTime() time.Time {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.time" {
				if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
					return t
				}
				return time.UnixMicro(0)
			}
		}
		return time.UnixMicro(0)
	}
	return time.UnixMicro(0)
}

// Print build time (or Unix epoch when build time is not set) and exit the program
func PrintBuildTime(ctx context.Context, cmd *cli.Command, b bool) error {
	if b {
		fmt.Println(getBuildTime().Unix())
		os.Exit(0)
	}
	return nil
}
