package logging

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/slugger/mstdnlambda/internal/cfg"
)

type LogCategory int

const (
	DefaultCategory LogCategory = iota
	DevEnvCategory
	DevEnvNotificationCategory
	HttpCategory
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
	case HttpCategory:
		return "http"
	case SnsNotificationCategory:
		return "SnsNotify"
	default:
		return "default"
	}
}

var Log *log.Entry

func init() {
	Reset()
}

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

func GetLogForCategory(cat LogCategory) *log.Entry {
	return Log.WithField("category", cat.String())
}

func AddField(key string, val interface{}) *log.Entry {
	Log = Log.WithField(key, val)
	return Log
}

func AddFields(fields log.Fields) *log.Entry {
	Log = Log.WithFields(fields)
	return Log
}

func LogAsJson(entry *log.Entry, lvl log.Level, subject interface{}, msg string, skipIfDebug bool) {
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
