package logging

/*
	mstdnlambda
	Copyright (C) 2022 Battams, Derek <derek@battams.ca>

	This program is free software; you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation; either version 2 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License along
	with this program; if not, write to the Free Software Foundation, Inc.,
	51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/cfg"
)

// LogCategory represents a known logging category type
type LogCategory int

// Supported logging categories
const (
	DefaultCategory LogCategory = iota
	DevEnvCategory
	DevEnvNotificationCategory
	HTTPCategory
	LambdaCategory
	SnsNotificationCategory
)

func (c LogCategory) String() string {
	switch c {
	case DevEnvCategory:
		return "DevEnv"
	case DevEnvNotificationCategory:
		return "DevEnvNotify"
	case LambdaCategory:
		return "lambda"
	case HTTPCategory:
		return "http"
	case SnsNotificationCategory:
		return "SnsNotify"
	default:
		return "default"
	}
}

// Log is the global log entry; preconfigured based on lambda configuration options
var Log *log.Entry

func init() {
	Reset()
}

// Reset puts the global logger back to its original state; should be called on each invocation of the lambda
func Reset() {
	l := log.New()
	l.SetFormatter(&log.JSONFormatter{})

	lvl, err := log.ParseLevel(cfg.Cfg.LogLevel())
	if err == nil {
		l.SetLevel(lvl)
	} else {
		l.SetLevel(log.DebugLevel)
		l.Warnf("invalid log level set, debug assumed [%s]", cfg.Cfg.LogLevel())
	}
	Log = log.NewEntry(l)
}

// GetLogForCategory returns a configured log.Entry for the given category
func GetLogForCategory(cat LogCategory) *log.Entry {
	return Log.WithField("category", cat.String())
}

// AddField allows adding a structured field and value to the global log entry; it does not add the field to the global entry itself but instead returns a new Entry with the field added
func AddField(key string, val interface{}) *log.Entry {
	Log = Log.WithField(key, val)
	return Log
}

// AddFields allows for the adding of multiple structure fields to the global log.Entry
func AddFields(fields log.Fields) *log.Entry {
	Log = Log.WithFields(fields)
	return Log
}

// LogAsJSON is a convienence method to easily marshal the subject to a json string and log that value in the structured field named "subject"
func LogAsJSON(entry *log.Entry, lvl log.Level, subject interface{}, msg string, skipIfDebug bool) {
	if !entry.Logger.IsLevelEnabled(lvl) || (skipIfDebug && entry.Logger.IsLevelEnabled(log.DebugLevel)) {
		return
	}

	enc, err := json.Marshal(subject)
	if err == nil {
		entry.WithField("subject", string(enc)).Log(lvl, msg)
	} else {
		Log.WithField("err", err).Error("json marshal failed")
	}
}
