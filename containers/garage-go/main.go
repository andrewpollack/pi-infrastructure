package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	PIN_NUMBER    = 12
	SLEEP_SECONDS = 2
	OPEN          = "Open"
	CLOSED        = "Closed"
)

var isGarageOpenGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "garage_door_open",
	Help: "Indicates whether the garage door is open (1 for open, 0 for closed).",
})

type state rpio.State

type Garage struct {
	isOpen bool
}

func (g *Garage) Open() {
	g.isOpen = true
}

func (g *Garage) Close() {
	g.isOpen = false
}

func (g *Garage) IsOpen() bool {
	return g.isOpen
}

func (s state) String() string {
	var currState string

	switch s {
	case state(rpio.Low):
		currState = CLOSED
	default:
		currState = OPEN
	}

	return currState
}

func monitorDoor(pin rpio.Pin) {
	for {
		rpio.PullMode(pin, rpio.PullUp)
		door := state(pin.Read())

		switch door.String() {
		case OPEN:
			isGarageOpenGauge.Set(1)
		case CLOSED:
			isGarageOpenGauge.Set(0)
		}

		// fmt.Printf("Door is: %s\n", door)
		time.Sleep(SLEEP_SECONDS * time.Second)
	}
}

func main() {
	// Create a custom Prometheus registry to avoid exporting Go metrics,
	// minimizing footprint.
	registry := prometheus.NewRegistry()
	registry.MustRegister(isGarageOpenGauge)

	// GPIO setup
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	pin := rpio.Pin(PIN_NUMBER)
	pin.Input()

	// Monitoring the door in a separate goroutine
	go monitorDoor(pin)

	// Start Prometheus HTTP server
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fmt.Println("Prometheus server running at :2135/metrics")
	if err := http.ListenAndServe(":2135", nil); err != nil {
		fmt.Printf("Error starting HTTP server: %v\n", err)
		os.Exit(1)
	}
}
