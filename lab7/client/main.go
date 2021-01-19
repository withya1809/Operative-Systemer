package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

var (
	serverAddr = flag.String("raddr", "192.168.0.10:61111", "address:port of the remote UDP server.")
	delay      = flag.Uint("delay", 3000, "Delay (in milliseconds) between each message sent to the server.")
	textfile   = flag.String("src", "loremipsum.txt", "Name of the textfile to use as the source for messages to the server.")
)

var sentences []string // sentences from the input file

func init() {
	// set the seed to prevent getting the same "random" sequence of sentences every time
	rand.Seed(time.Now().UnixNano())
}

// readSource reads the input file and stores each line in `sentences`
func readSource() error {
	log.Printf("Reading '%s'...", *textfile)

	f, err := os.Open(*textfile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if len(s) > 1 { // ignore empty lines
			sentences = append(sentences, s)
		}

	}
	return scanner.Err()
}

// setupConnection tries to parse the remote address and initialize a `*net.UDPConn` object.
func setupConnection() (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr("udp", *serverAddr)
	if err != nil {
		return nil, fmt.Errorf("address translation failed: %w", err)
	}

	log.Printf("Dialing '%s'...", *serverAddr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial remote server: %w", err)
	}

	return conn, nil
}

// sendAndReceive continuously picks a random sentence from the input file,
// sends it to the remote server at `conn` and reads and prints the result.
func sendAndReceive(conn *net.UDPConn) error {
	log.Println("Starting to message server...")

	t := time.Tick(time.Duration(*delay) * time.Millisecond)
	for range t {
		// pick a random sentence and send it to the server
		s := sentences[rand.Intn(len(sentences))]
		_, err := conn.Write([]byte(s))
		if err != nil {
			return fmt.Errorf("failed to write to the server: %w", err)
		}

		// read the result from the server and print it
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			return fmt.Errorf("failed to read from the server: %w", err)
		}
		log.Println(string(b[:n]))
	}

	return nil
}

func main() {
	flag.Parse()

	err := readSource()
	if err != nil {
		panic(fmt.Errorf("failed to process input file: %v", err))
	}

	conn, err := setupConnection()
	if err != nil {
		panic(err)
	}

	err = sendAndReceive(conn)
	conn.Close()

	if err != nil {
		log.Printf("stopped processing due to error: %v", err)
	}
}
