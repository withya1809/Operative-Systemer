package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var (
	hostName = flag.String("addr", ":61112", "address:port to listen for TCP requests.")
	interval = flag.Uint("t", 10, "How often (in seconds) the logger prints its current state.")
)

// logger logs the number of messages received by hosts connected to a remote server
type logger struct {
	log      map[string]int
	listener *net.TCPListener
	lock     sync.Mutex
}

// msg is used by the JSON-formatted message sent by the server
type msg struct {
	Client string
}

func newLogger() *logger {
	return &logger{log: make(map[string]int)}
}

// updateLog increments the client's counter
func (l *logger) updateLog(client string) {
	l.lock.Lock()
	l.log[client]++
	l.lock.Unlock()
}

// startListener sets up a TCP listener for the logger
func (l *logger) startListener() error {
	log.Println("Setting up listener...")

	addr, err := net.ResolveTCPAddr("tcp", *hostName)
	if err != nil {
		return fmt.Errorf("address translation failed: %w", err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to setup TCP listener: %w", err)
	}

	l.listener = listener
	return nil
}

// handle accepts TCP connections and handles requests from them.
// It also prints the state of the logger at a regular interval.
func (l *logger) handle() {
	// print state in the background
	go func() {
		t := time.Tick(time.Duration(*interval) * time.Second)
		for range t {
			l.printState()
		}
	}()

	// accept TCP connections and handle requests in the background
	for {
		conn, err := l.listener.AcceptTCP()
		if err != nil {
			log.Printf("Failed connecting to %v: %v", conn.RemoteAddr(), err)
			continue
		}
		log.Printf("Connected to '%v'...", conn.RemoteAddr())

		// handle requests in the background
		go func(conn *net.TCPConn) {
			defer conn.Close()
			b := make([]byte, 1024)
			for {
				n, err := conn.Read(b)
				if err != nil {
					log.Printf("Failed reading from %v. Closing connection. Error: %v", conn.RemoteAddr(), err)
					return
				}

				// translate the JSON-formatted message
				m := &msg{}
				err = json.Unmarshal(b[:n], m)
				if err != nil {
					log.Printf("Failed to unmarshal message from %v. Ignoring the message. Error: %v", conn.RemoteAddr(), err)
					continue
				}

				// update the log
				l.updateLog(m.Client)
			}
		}(conn)
	}
}

// printState prints the current internal state of the log
func (l *logger) printState() {
	l.lock.Lock()
	defer l.lock.Unlock()

	var str string
	for k, v := range l.log {
		str += fmt.Sprintf("(%s: %d)", k, v)
	}
	log.Println("Current state:", str)
}

func main() {
	l := newLogger()
	err := l.startListener()
	defer l.listener.Close()
	if err != nil {
		panic(err)
	}
	l.handle()
}
