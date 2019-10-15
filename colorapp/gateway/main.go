package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
)

const (
	defaultPort = "8080"
	xrayDefaultLogLevel = "warn"
)

var (
	enableXrayTracing bool
)

var (
	errorLog *log.Logger
	infoLog  *log.Logger
	pingLog *log.Logger
)

func init() {
	errorLog = log.New(os.Stderr, "[ERROR] ", 0)
	infoLog = log.New(os.Stdout, "[INFO] ", 0)
	pingLog = log.New(os.Stdout, "[PING] ", 0)

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
}

// Set SERVER_PORT environment variable to override the default listen port
func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

// COLOR_TELLER_ENDPOINT environment value must be set
// ex: colorteller:8080/mesh.local
func getColorTellerEndpoint() (string, error) {
	colorTellerEndpoint := os.Getenv("COLOR_TELLER_ENDPOINT")
	if colorTellerEndpoint == "" {
		return "", errors.New("COLOR_TELLER_ENDPOINT is not set")
	}
	return colorTellerEndpoint, nil
}

// Fetch a color from the color teller endpoint
func getColorFromColorTeller(request *http.Request) (string, error) {
	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s", colorTellerEndpoint), nil)
	if err != nil {
		return "", err
	}

	var client *http.Client
	if enableXrayTracing {
		client = xray.Client(&http.Client{})
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req.WithContext(request.Context()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// colorteller "red" is configured to occasionally return 500
	// in any case, client.Do doesn't consider HTTP error codes to be
	// an error, so let's handle that and plug the gap in error
	// handling that allows errors to slip through as colors and get
	// stored in the color history.
	if resp.StatusCode >= 400 {
		return "", errors.New(string(body))
	}

	color := strings.TrimSpace(string(body))
	if len(color) < 1 {
		return "", errors.New("empty response from colorteller")
	}

	return color, nil
}

// /color

type colorHandler struct{}

func (h *colorHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ep, _ := getColorTellerEndpoint()
	infoLog.Printf("Requesting color from: %s", ep)

	color, err := getColorFromColorTeller(request)
	if err != nil {
		errorLog.Printf("Request for color failed: %s", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("500 - Internal Error"))
		return
	}

	infoLog.Printf("Request for color fetched: %s", color)
	addColor(color)

	statsJson, err := json.Marshal(getRatios())
	if err != nil {
		errorLog.Printf("Failed to get color stats: %s", err)
		fmt.Fprintf(writer, `{"color":"%s", "error":"%s"}`, color, err)
		return
	}
	infoLog.Printf("Sending color response: {\"color\":\"%s\", \"stats\":%s}", color, statsJson)
	fmt.Fprintf(writer, `{"color":"%s", "stats": %s}`, color, statsJson)
}

// /stats/clear

type clearColorStatsHandler struct{}

func (h *clearColorStatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	clearColors()
	infoLog.Print("Cleared color stats")
	fmt.Fprint(writer, "Cleared color stats")
}

// /stats

type getColorStatsHandler struct{}

func (h *getColorStatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	statsJson, err := json.Marshal(getRatios())
	if err != nil {
		errorLog.Printf("Failed to get color stats: %s", err)
		fmt.Fprintf(writer, "Failed to get color stats: %s}", err)
		return
	}
	infoLog.Printf("Sending stats response: {\"stats\":%s}", statsJson)
	fmt.Fprintf(writer, `{"stats": %s}`, statsJson)
}

// /ping

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	pingLog.Println("Responding to ping with HTTP 200")
	writer.WriteHeader(http.StatusOK)
}

func main() {
	infoLog.Printf("Starting gateway on port: %s", getServerPort())

	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		log.Fatalln(err)
	}
	infoLog.Println("Using colorteller at: " + colorTellerEndpoint)

	var color http.Handler
	var clear http.Handler
	var stats http.Handler
	var ping http.Handler

	if enableXrayTracing {
		xraySegmentNamer := xray.NewFixedSegmentNamer("gateway")
		color = xray.Handler(xraySegmentNamer, &colorHandler{})
		clear = xray.Handler(xraySegmentNamer, &clearColorStatsHandler{})
		stats = xray.Handler(xraySegmentNamer, &getColorStatsHandler{})
		ping = xray.Handler(xraySegmentNamer, &pingHandler{})
	} else {
		color = &colorHandler{}
		clear = &clearColorStatsHandler{}
		stats = &getColorStatsHandler{}
		ping = &pingHandler{}
	}

	// leave /color/clear for existing clients
	http.Handle("/color/clear", clear)
	http.Handle("/color", color)
	http.Handle("/stats/clear", clear)
	http.Handle("/stats", stats)
	http.Handle("/ping", ping)

	log.Fatal(http.ListenAndServe(":"+getServerPort(), nil))
}
