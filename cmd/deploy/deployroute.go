package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

const (
	appKey = "APP"
)

var (
	errorLogger *log.Logger
	infoLogger  *log.Logger

	appUrl *url.URL
)

func init() {
	errorLogger = log.New(os.Stderr, "[ERROR] ", 0)
	infoLogger = log.New(os.Stdout, "[INFO] ", 0)
}

func print(fmtstr string, fmtArgs ...interface{}) {
	fmt.Printf(fmtstr, fmtArgs...)
}

func info(fmtstr string, fmtArgs ...interface{}) {
	infoLogger.Printf(fmtstr, fmtArgs...)
}

func error(fmtstr string, fmtArgs ...interface{}) {
	errorLogger.Printf(fmtstr, fmtArgs...)
}

func fatal(exitCode int, fmtstr string, fmtArgs ...interface{}) {
	error(fmtstr, fmtArgs...)
	os.Exit(exitCode)
}

func readEnvVars() {
	if s := os.Getenv(appKey); s == "" {
		fatal(1, "environment variable $APP should be set to a URL for the app")
	} else {
		if u, err := url.Parse(s); err != nil {
			fatal(1, "not a valid url for $APP (%s): %s", s, err)
		} else {
			appUrl = u
		}
	}
}

func main() {
	readEnvVars()
	info(appUrl.String())
}
