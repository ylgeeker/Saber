/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package logger

import (
	"log"

	"go.uber.org/zap"
)

// Logger is a universal logging interface that can be
// flexibly replaced with other logging libraies
// without interfering with the operational logic
// of business code.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

type Level string

var (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
)

// Config logger config
type Config struct {
	Filename   string
	LogLevel   Level
	MaxSizeMB  int
	MaxBackups int
	MaxAge     int
}

var l Logger

func SetLogger(log Logger) {
	l = log
}

func GetOriginLogger() *zap.Logger {
	if l == nil {
		return nil
	}

	return l.(*ZapLogger).logger
}

func Debugf(format string, args ...any) {
	if l == nil {
		log.Printf(format, args...)
		return
	}

	l.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	if l == nil {
		log.Printf(format, args...)
		return
	}

	l.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	if l == nil {
		log.Printf(format, args...)
		return
	}

	l.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	if l == nil {
		log.Printf(format, args...)
		return
	}

	l.Errorf(format, args...)
}

func Fatalf(format string, args ...any) {
	if l == nil {
		log.Fatalf(format, args...)
		return
	}

	l.Fatalf(format, args...)
}
