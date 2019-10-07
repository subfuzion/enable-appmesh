package main

import (
	"math"
	"sync"
)

const maxColors = 1000

var colors [maxColors]string
var colorsIdx int
var colorsMutex = &sync.Mutex{}

func addColor(color string) {
	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	colors[colorsIdx] = color
	colorsIdx += 1
	if colorsIdx >= maxColors {
		colorsIdx = 0
	}
}

func getRatios() map[string]float64 {
	counts := make(map[string]int)
	var total = 0

	for _, c := range colors {
		if c != "" {
			counts[c] += 1
			total += 1
		}
	}

	ratios := make(map[string]float64)
	for k, v := range counts {
		ratio := float64(v) / float64(total)
		ratios[k] = math.Round(ratio*100) / 100
	}

	return ratios
}

func clearColors() {
	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	colorsIdx = 0

	for i := range colors {
		colors[i] = ""
	}
}
