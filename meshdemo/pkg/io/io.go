/*
Copyright © 2019 Tony Pujals <tpujals@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Package io provides consistent input and output functionality for the demo CLI tools.
package io

import (
	"fmt"
	"log"
	"os"
)

// only tested on a mac so far, but alignment spacing renders correctly
const (
	failPrefix    = "❌ [FAILURE] "
	successPrefix = "✅ [SUCCESS] "
	warnPrefix    = "⚠️  [CAUTION] "
	alertPrefix   = "❗ [ALERT  ] "
	statusPrefix  = " ✔︎ "
	stepPrefix    = " ▷ "
)

var (
	errorLogger *log.Logger
	infoLogger  *log.Logger
)

func init() {
	errorLogger = log.New(os.Stderr, failPrefix, 0)
	infoLogger = log.New(os.Stdout, "", 0)
}

// Printf prints formatted string to stdout
func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Println prints formatted string to stdout and adds a newline
func Println(format string, a ...interface{}) {
	Printf(format, a...)
	Printf("\n")
}

// Info prints formatted string to stdout using internal logger
func Info(format string, a ...interface{}) {
	infoLogger.Printf(format, a...)
}

// Error prints formatted error to stderr using internal logger
// format is an interface instead of string so the function also accepts errors
func Error(format interface{}, a ...interface{}) {
	errorLogger.Printf(fmt.Sprintf("%s", format), a...)
}

// Fatal prints formatted error to stderr and exits with an error code
// format is an interface instead of string so the function also accepts errors
func Fatal(exitCode int, format interface{}, a ...interface{}) {
	Error(format, a...)
	os.Exit(exitCode)
}

// Success should be used to print successful outcome / status
func Success(format string, a ...interface{}) {
	Printf(successPrefix)
	Println(format, a...)
}

// Failed should be used to print failure outcome / status
func Failed(format string, a ...interface{}) {
	Printf(failPrefix)
	Println(format, a...)
}

// Warn should be used to print caution notices
func Warn(format string, a ...interface{}) {
	Printf(warnPrefix)
	Println(format, a...)
}

// Alert should be used to print important notices
func Alert(format string, a ...interface{}) {
	Printf(alertPrefix)
	Println(format, a...)
}

// Step should be used to print initiating each step in a sequence of steps or operations
func Step(format string, a ...interface{}) {
	Printf(stepPrefix)
	Println(format, a...)
}

// Status should be used to print (successful) status updates during a sequence of steps or long operation
func Status(format string, a ...interface{}) {
	Printf(statusPrefix)
	Println(format, a...)
}
