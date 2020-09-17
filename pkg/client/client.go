package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"time"
)

// Message for backend
type Message struct {
	Price    int `json:"price"`
	Quantity int `json:"quantity"`
	Amount   int `json:"amount"`
	Object   int `json:"object"`
	Method   int `json:"method"`
}

// Client for backend
type Client struct {
	Protocol string
	Host     string
	Port     string
	Sended   uint
	Success  uint
}

// Send send messages to backend
func (client *Client) Send() {
	conn, err := net.Dial(client.Protocol, fmt.Sprintf("%v:%v", client.Host, client.Port))
	if err != nil {
		fmt.Println("Host refused connection.")
		return
	}
	fmt.Printf("%v Send package to %v://%v:%v\n", time.Now(), client.Protocol, client.Host, client.Port)
	go client.writeMessage(conn)
}

// writeMessage write message to connection
func (client *Client) writeMessage(conn net.Conn) {
	defer conn.Close()

	message, err := createMessageJSON()
	if err != nil {
		fmt.Printf("Can`t create message, error: %v\n", err)
		return
	}
	client.Sended++

	fmt.Fprintf(conn, string(message))
	answer, errRead := ioutil.ReadAll(conn)
	if errRead != nil {
		fmt.Println(errRead)
	}
	if string(answer) == "OK" {
		client.Success++
		fmt.Println(client.Sended, client.Success)
	}
}

// createMessageJSON create random JSON message
func createMessageJSON() ([]byte, error) {
	messages := []Message{}
	for i := 0; i <= rand.Intn(10); i++ {
		messages = append(messages, Message{
			Price:    rand.Intn(500),
			Quantity: rand.Intn(500),
			Amount:   rand.Intn(500),
			Object:   rand.Intn(500),
			Method:   rand.Intn(500),
		})
	}
	return json.Marshal(messages)
}
