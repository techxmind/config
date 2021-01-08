package config

import (
	_logger "github.com/techxmind/logger"
)

type Logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})
}

var (
	logger Logger = _logger.Named("config")
)

func SetLogger(l Logger) {
	logger = l
}
