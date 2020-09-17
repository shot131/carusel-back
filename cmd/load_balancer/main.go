package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"../../pkg/backend"
)

const (
	maxBackends  = 3
	defaultSpeed = 3
	logFile      = "../../server.log"
)

var limits *string

func init() {
	limits = flag.String("limits", fmt.Sprint(defaultSpeed), "messages receive limit per second")
}

func main() {
	flag.Parse()
	logFilePtr, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFilePtr.Close()
	initLog(logFilePtr)

	port := 8081
	backends := new([maxBackends]*backend.Backend)
	speedLimits := getLimits(*limits)
	for i := 0; i < maxBackends; i++ {
		speed, err := strconv.Atoi(speedLimits[i])
		if err != nil {
			panic("Limits must be digits")
		}
		backends[i] = &backend.Backend{
			ID:       uint(i),
			Speed:    uint(speed),
			Protocol: "tcp",
			Host:     "localhost",
			Port:     port,
		}
		port++
		if i > 0 {
			backends[i-1].NextBackend = backends[i]
		}
	}

	for _, backend := range backends {
		go backend.Start()
	}

	fmt.Println("Press enter for exit")
	fmt.Scanln()
}

func initLog(file *os.File) {
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func getLimits(limits string) []string {
	speedLimits := strings.Split(limits, ",")[:]
	if len(speedLimits) < maxBackends {
		for i := 0; i < maxBackends; i++ {
			if i > len(speedLimits)-1 {
				speedLimits = append(speedLimits, fmt.Sprint(defaultSpeed))
			}
		}
	}
	return speedLimits
}
