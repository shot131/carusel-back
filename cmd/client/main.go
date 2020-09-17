package main

import (
	"flag"
	"fmt"
	"time"

	"../../pkg/client"
)

var sendSpeed *uint

func init() {
	sendSpeed = flag.Uint("speed", 1, "messages send speed per second")
}

func main() {
	flag.Parse()
	client := client.Client{
		Protocol: "tcp",
		Host:     "localhost",
		Port:     "8081",
	}

	closeClient := make(chan bool)
	go func() {
		for {
			select {
			case <-closeClient:
				break
			default:
				go client.Send()
				time.Sleep(time.Second / time.Duration(*sendSpeed))
			}
		}
	}()

	fmt.Println("Press enter for exit")
	fmt.Scanln()
	closeClient <- true
	fmt.Printf("Sended: %v\n", client.Sended)
	fmt.Printf("Success: %v\n", client.Success)
}
