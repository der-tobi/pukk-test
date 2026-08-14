package main

import "log"

type standardLogger struct{}

func (standardLogger) Printf(format string, args ...any) {
	log.Printf(format, args...)
}
