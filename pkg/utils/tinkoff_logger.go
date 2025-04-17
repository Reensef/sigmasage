package utils

import (
	"log"
)

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
