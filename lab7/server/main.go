package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strings"
	"unicode"
)

var (
	hostName   = flag.String("addr", ":61111", "address:port to listen for incoming UDP messages.")
	loggerAddr = flag.String("logaddr", "192.168.0.20:61112", "address:port of the logger.")
)

// setupListener sets up a UDP listener at `hostName`
func setupListener() (*net.UDPConn, error) {
	log.Println("Setting up listener...")

	addr, err := net.ResolveUDPAddr("udp", *hostName)
	if err != nil {
		return nil, fmt.Errorf("address translation failed: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to setup UDP listener: %w", err)
	}

	return conn, nil
}

// listen continuously receives messages from the listener, processes the
// message, and sends it back to the sender.
func listen(conn *net.UDPConn, logger *net.TCPConn) error {
	defer conn.Close()

	log.Printf("Starting to listen at '%s'...", conn.LocalAddr())
	for {
		b := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("connection closed: %v", err)
			}
			log.Printf("failed reading from %v: %v", addr, err)
			continue
		}

		// if logger is not present, the server prints some information itself
		if logger != nil {
			go writeToLog(logger, addr.String())
		} else {
			log.Printf("Processing message from '%v'.", addr)
		}
		go handleAndRespond(b[:n], conn, addr)

	}
}

// handleAndRespond converts `seq` and sends the result back to `addr` over `conn`
func handleAndRespond(seq []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	// randomly set the case of each letter
	res := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			if rand.Intn(2) == 1 {
				return unicode.ToUpper(r)
			}
		}
		return r
	}, string(seq))

	_, err := conn.WriteToUDP([]byte(res), addr)
	if err != nil {
		log.Printf("failed writing to %v: %v", addr, err)
	}
}

// setupLogger tries to setup a TCP connection to the logger
func setupLogger() (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", *loggerAddr)
	if err != nil {
		return nil, fmt.Errorf("address translation failed: %w", err)
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to logger: %w", err)
	}
	log.Printf("Connected to logger at %v", conn.RemoteAddr())

	return conn, nil
}

type msg struct {
	Client string
}

// writeToLog messages the log about an event from `addr`
func writeToLog(logger *net.TCPConn, addr string) {
	// parse to JSON
	msg := msg{addr}
	m, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal message. Error: %v", err)
		return
	}

	logger.Write(m)
	if err != nil {
		log.Printf("Failed writing to the logger. Error: %v", err)
	}
}

func main() {
	flag.Parse()

	conn, err := setupListener()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	logger, err := setupLogger()
	if err != nil {
		log.Printf("Logger will not be used: %v", err)
	} else {
		defer logger.Close()
	}

	err = listen(conn, logger)
	if err != nil {
		log.Printf("stopped processing due to error: %v", err)
	}
}
