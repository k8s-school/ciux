// Copyright 2019 The go-daq Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log provides routines for logging messages.
package log // import "github.com/go-daq/tdaq/log"

import (
	"fmt"
	"log"
	"os"
	"strings"
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

// String prints the human-readable representation of a Level value.
func (lvl Level) String() string {
	switch lvl {
	case LvlDebug:
		return "DEBUG"
	case LvlInfo:
		return "INFO"
	case LvlWarning:
		return "WARN"
	case LvlError:
		return "ERROR"
	}
	panic(fmt.Errorf("log: invalid log.Level value [%d]", int(lvl)))
}

// MsgStream provides access to verbosity-defined formated messages, a la fmt.Printf.
type MsgStream interface {
	Printf(lvl Level, format string, a ...interface{})
}

type msgstream struct {
	lvl    Level
	prefix string
}

var (
	msgStream = newMsgStream("ciux", LvlDebug)
)

func SetLogLevel(verbosity int) {
	var lvl Level

	switch verbosity {
	case 0:
		lvl = LvlError
	case 1:
		lvl = LvlInfo
	case 2:
		lvl = LvlDebug
	default:
		lvl = LvlDebug
	}

	msgStream.lvl = lvl
}

// Debugf displays a (formated) DBG message
func Debugf(format string, a ...interface{}) {
	msgStream.Printf(LvlDebug, format, a...)
}

// Infof displays a (formated) INFO message
func Infof(format string, a ...interface{}) {
	msgStream.Printf(LvlInfo, format, a...)
}

// Warnf displays a (formated) WARN message
func Warnf(format string, a ...interface{}) {
	msgStream.Printf(LvlWarning, format, a...)
}

// Errorf displays a (formated) ERR message
func Errorf(format string, a ...interface{}) {
	msgStream.Printf(LvlError, format, a...)
}

// Fatalf displays a (formated) ERR message and stops the program.
func Fatalf(format string, a ...interface{}) {
	msgStream.Printf(LvlError, format, a...)
	os.Exit(1)
}

// Panicf displays a (formated) ERR message and panics.
func Panicf(format string, a ...interface{}) {
	msgStream.Printf(LvlError, format, a...)
	panic("ciux panic")
}

func newMsgStream(name string, lvl Level) msgstream {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return msgstream{
		lvl:    lvl,
		prefix: fmt.Sprintf("%-2s ", name),
	}
}

// Msg displays a (formated) message with level lvl.
func (msg msgstream) Printf(lvl Level, format string, a ...interface{}) {
	if lvl < msg.lvl {
		return
	}
	format = msg.prefix + lvl.String() + " " + format
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	log.Printf(format, a...)
}
