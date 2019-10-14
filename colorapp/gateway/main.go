package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPort = "8080"
	xrayDefaultLogLevel = "warn"
)

var (
	enableXrayTracing bool
)

func init() {
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
		return "", errors.New("[Error] COLOR_TELLER_ENDPOINT is not set")
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
	log.Printf("[Info] fetching color from: %s", ep)

	color, err := getColorFromColorTeller(request)
	if err != nil {
		log.Printf("[Error] fetching color (%s)", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("500 - Internal Error"))
		return
	}

	log.Printf("[Info] fetched color: %s", color)
	addColor(color)

	statsJson, err := json.Marshal(getRatios())
	if err != nil {
		log.Printf("[Error] getting color stats: %s", err)
		fmt.Fprintf(writer, `{"color":"%s", "error":"%s"}`, color, err)
		return
	}
	log.Printf("[Info] sending response: {\"color\":\"%s\", \"stats\":%s}", color, statsJson)
	fmt.Fprintf(writer, `{"color":"%s", "stats": %s}`, color, statsJson)
}

// /color/clear

type clearColorStatsHandler struct{}

func (h *clearColorStatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	clearColors()
	log.Print("[Info] cleared color stats")
	fmt.Fprint(writer, "cleared")
}

// /ping

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("[Ping] responding with HTTP 200")
	writer.WriteHeader(http.StatusOK)
}

func main() {
	log.Printf("[Info] Starting gateway, listening on port %s", getServerPort())

	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("[Info] Using colorteller at " + colorTellerEndpoint)

	var color http.Handler
	var clear http.Handler
	var ping http.Handler

	if enableXrayTracing {
		xraySegmentNamer := xray.NewFixedSegmentNamer("gateway")
		color = xray.Handler(xraySegmentNamer, &colorHandler{})
		clear = xray.Handler(xraySegmentNamer, &clearColorStatsHandler{})
		ping = xray.Handler(xraySegmentNamer, &pingHandler{})
	} else {
		color = &colorHandler{}
		clear = &clearColorStatsHandler{}
		ping = &pingHandler{}
	}
	http.Handle("/color", color)
	http.Handle("/color/clear", clear)
	http.Handle("/ping", ping)

	log.Fatal(http.ListenAndServe(":"+getServerPort(), nil))
}
