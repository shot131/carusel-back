package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"
)

// Backend for messages logging
type Backend struct {
	ID          uint
	Speed       uint
	Protocol    string
	Host        string
	Port        int
	NextBackend *Backend
	logChan     chan string
	clientChan  chan []byte
	receiveTime time.Time
	received    uint
}

// Start start listener and handle connections
func (backend *Backend) Start() {
	backend.received = 0
	backend.logChan = make(chan string)
	backend.clientChan = make(chan []byte)

	listener, err := net.Listen(backend.Protocol, fmt.Sprintf("%v:%v", backend.Host, backend.Port))
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer listener.Close()

	fmt.Printf("Listening on %v://%v:%v, speed: %v\n", backend.Protocol, backend.Host, backend.Port, backend.Speed)

	for {
		conn, err := listener.Accept()
		if err != nil {
			backend.logError(err)
			continue
		}
		go backend.closeClient(conn)
		go backend.handleRequest(conn)
		go backend.log()
	}
}

// handleRequest handle request or proxy to next backend
func (backend *Backend) handleRequest(conn net.Conn) {

	if time.Since(backend.receiveTime) < time.Second*1 {
		backend.received++
	} else {
		backend.received = 1
		backend.receiveTime = time.Now()
	}

	if backend.received > backend.Speed {
		go backend.proxyRequest(conn)
		return
	}

	message, err := decodeFromJSON(conn)
	if err != nil {
		backend.logError(err)
		return
	}

	backend.clientChan <- []byte("OK")
	backend.logChan <- fmt.Sprintf("message %v", message)
}

// proxyRequest proxy request to next backend
func (backend *Backend) proxyRequest(conn net.Conn) {
	if backend.NextBackend == nil {
		message := "error: can`t proceed request"
		backend.clientChan <- []byte(message)
		backend.logChan <- message
		return
	}
	backend.logChan <- fmt.Sprintf("proxy request to backend: %v", backend.NextBackend.ID)

	remoteBackend := backend.NextBackend
	remoteConn, err := net.Dial(remoteBackend.Protocol, fmt.Sprintf("%v:%v", remoteBackend.Host, remoteBackend.Port))
	if err != nil {
		backend.logError(err)
		return
	}

	go backend.writeProxyRequest(conn, remoteConn)
	go backend.readProxyResponse(remoteConn)
}

// writeProxyRequest write request to next backend connection
func (backend *Backend) writeProxyRequest(src net.Conn, dest net.Conn) {
	message, err := ioutil.ReadAll(src)
	if err != nil {
		backend.logError(err)
	}
	if _, err := dest.Write(message); err != nil {
		backend.logError(err)
	}
}

// readProxyRequest read answer from next backend connection
func (backend *Backend) readProxyResponse(dest net.Conn) {
	defer dest.Close()
	answer, err := ioutil.ReadAll(dest)
	if err != nil {
		backend.logError(err)
	}
	backend.clientChan <- answer
}

// closeClient close client connection and send status message
func (backend *Backend) closeClient(conn net.Conn) {
	defer conn.Close()
	_, err := conn.Write(<-backend.clientChan)
	if err != nil {
		backend.logError(err)
	}
}

func (backend *Backend) log() {
	log.Println(backend.formatMsg(<-backend.logChan))
}

func (backend *Backend) logError(err error) {
	log.Println(backend.formatMsg(fmt.Sprintf("error: %v", err)))
}

func (backend *Backend) formatMsg(message string) string {
	return fmt.Sprintf(
		"backend %v - %v",
		backend.ID,
		message,
	)
}

func decodeFromJSON(conn net.Conn) ([]map[string]int, error) {
	jsonDecoder := json.NewDecoder(conn)
	var message []map[string]int
	err := jsonDecoder.Decode(&message)
	return message, err
}
