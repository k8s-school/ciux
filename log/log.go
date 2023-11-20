// Copyright 2019 The go-daq Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log provides routines for logging messages.
package log // import "github.com/go-daq/tdaq/log"

import (
	"context"
	"log/slog"
	"os"
)

// Level regulates the verbosity level of a component.
type Level int

// Default verbosity levels.
const (
	LvlDebug   Level = -10 // LvlDebug defines the DBG verbosity level
	LvlInfo    Level = 0   // LvlInfo defines the INFO verbosity level
	LvlWarning Level = 10  // LvlWarning defines the WARN verbosity level
	LvlError   Level = 20  // LvlError defines the ERR verbosity level
)

func SetLogLevel(verbosity int) {
	var lvl slog.Leveler

	switch verbosity {
	case 0:
		lvl = slog.LevelError
	case 1:
		lvl = slog.LevelInfo
	case 2:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func IsDebugEnabled() bool {
	h := slog.Default().Handler()
	return h.Enabled(context.Background(), slog.LevelDebug)
}
