package utils

import (
	"log"
)

// TODO Need rename
// This is a implementation of logger for tinkoff api
// He has methods of uber zap logger
type TinkoffLogger struct {
}

func (l TinkoffLogger) Infof(template string, args ...any) {
	log.Printf(template, args...)
}

func (l TinkoffLogger) Errorf(template string, args ...any) {
	log.Printf(template, args...)
}

func (l TinkoffLogger) Fatalf(template string, args ...any) {
	log.Fatalf(template, args...)
}
