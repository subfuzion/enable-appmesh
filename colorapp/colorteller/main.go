package main

import (
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	defaultPort = "8080"
	defaultColor = "black"
	xrayDefaultLogLevel = "warn"
)

var (
	enableXrayTracing bool

	// The following is to support configuring the service with faults
	// (until App Mesh supports fault injection)
	responseDelay time.Duration
	periodicError int
	counter int
)


func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// ENABLE_XRAY_TRACING
	if enable, err := strconv.ParseBool(os.Getenv("ENABLE_XRAY_TRACING")); err == nil {
		enableXrayTracing = enable
	}

	if enableXrayTracing {
		xrayLogLevel := os.Getenv("XRAY_LOG_LEVEL")
		if xrayLogLevel == "" {
			xrayLogLevel = xrayDefaultLogLevel
		}

		xray.Configure(xray.Config{
			LogLevel: xrayLogLevel,
		})
	}

	// TEST_RESPONSE_DELAY should be in ms
	// This is how long the main route will delay before sending a response
	// A zero value means no delay
	// don't report error when parse fails because env var is not set
	if delayStr, exists := os.LookupEnv("TEST_RESPONSE_DELAY"); exists {
		delay, err := strconv.Atoi(delayStr)
		if err != nil {
			log.Printf("[Error] failed to parse TEST_RESPONSE_DELAY (%s): %s", delayStr, err)
		}
		responseDelay = time.Duration(delay)
	}

	// TEST_PERIODIC_ERROR is a number that means to send an error (HTTP 500)
	// every so many invocations of the default route
	// A zero value means never send errors; a 1 value means send an error every
	// invocation; a 2 value means every other invocation; and so on.
	if peStr, exists := os.LookupEnv("TEST_PERIODIC_ERROR"); exists {
		pe, err := strconv.Atoi(peStr)
		if err != nil {
			log.Printf("[Error] failed to parse TEST_PERIODIC_ERROR (%s): %s", peStr, err)
		}
		periodicError = pe
	}

}

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

func getColor() string {
	color := os.Getenv("COLOR")
	if color != "" {
		return color
	}

	return defaultColor
}

type colorHandler struct{}
func (h *colorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Printf("[Info] color requested, responding with %s", getColor())

	if responseDelay > 0 {
		log.Printf("[Test] delaying response for %d ms", responseDelay)
		time.Sleep(time.Millisecond * responseDelay)
	}

	if periodicError > 0 {
		if counter >= periodicError {
			counter = 0
		}
		counter++
		log.Printf("[Test] counter = %d", counter)

		if counter == periodicError {
			log.Printf("[Test] sending 500 for color %s => periodic error = %d (counter = %d)", getColor(), periodicError, counter)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("[Test] periodic error (expected) for color %s => periodic=%d (counter=%d), sending HTTP 500", getColor(), periodicError, counter)))
			return
		}
	}

	fmt.Fprint(writer, getColor())
}

type pingHandler struct{}
func (h *pingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Print("[Ping] reponding with HTTP 200")
	writer.WriteHeader(http.StatusOK)
}

func main() {
	log.Printf("[Info] Starting colorteller (%s), listening on port %s", getColor(), getServerPort())
	if responseDelay > 0 {
		log.Printf("[Test] enabling response delays (delay = %d ms)", responseDelay)
	}
	if periodicError > 0 {
		log.Printf("[Test] enabling periodic request errors (period = %d)", periodicError)
	}

	var color http.Handler
	var ping http.Handler

	if enableXrayTracing {
		xraySegmentNamer := xray.NewFixedSegmentNamer(fmt.Sprintf("colorteller-%s", getColor()))
		color = xray.Handler(xraySegmentNamer, &colorHandler{})
		ping = xray.Handler(xraySegmentNamer, &pingHandler{})
	} else {
		color = &colorHandler{}
		ping = &pingHandler{}
	}
	http.Handle("/", color)
	http.Handle("/ping", ping)

	http.ListenAndServe(":"+getServerPort(), nil)
}
