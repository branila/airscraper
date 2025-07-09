package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

func decodeLZW(input string) string {
	// Initialize the dictionary: codes 0-255 (ASCII characters)
	dict := make(map[int]string)
	for i := range 256 {
		dict[i] = string(rune(i))
	}

	// State variables
	data := []rune(input)
	if len(data) == 0 {
		return ""
	}

	var result []rune
	prev := string(data[0])
	result = append(result, rune(data[0]))
	code := 256

	for i := 1; i < len(data); i++ {
		currCode := int(data[i])
		var entry string

		if currCode < 256 {
			entry = string(rune(currCode))
		} else if val, ok := dict[currCode]; ok {
			entry = val
		} else {
			// Special case: entry not yet in the dictionary
			entry = prev + string(prev[0])
		}

		// Add to the decompressed string
		result = append(result, []rune(entry)...)

		// Update the dictionary
		dict[code] = prev + string(entry[0])
		code++
		prev = entry
	}

	return string(result)
}

// Reads messages from the WebSocket connection
func readMessages(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		// Reads a message from the WebSocket connection
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("Unexpected WebSocket error: %v", err)
			}
			return
		}

		if messageType == websocket.CloseMessage {
			return
		}

		message = []byte(decodeLZW(string(message)))

		// Prettify output in json format
		var prettyMessage map[string]any
		if err := json.Unmarshal(message, &prettyMessage); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		prettyMessageJSON, err := json.MarshalIndent(prettyMessage, "", "  ")
		if err != nil {
			log.Printf("Error marshalling message to JSON: %v", err)
			continue
		}

		message = prettyMessageJSON

		fmt.Printf("Received message: %s\n", message)
	}
}

func main() {
	url := "wss://ws1.blitzortung.org/"

	fmt.Printf("Connecting to %s...\n", url)

	dialer := websocket.DefaultDialer

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	fmt.Println("Connection established with", url)

	err = conn.WriteMessage(websocket.TextMessage, []byte("{\"a\":111}"))
	if err != nil {
		log.Fatal("Error sending message:", err)
	}

	fmt.Println("Message sent successfully")

	// Channel for interrupt signals
	interrupt := make(chan os.Signal, 1)

	// Capture interrupt signals to gracefully close the connection
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Channel for done reading messages
	done := make(chan struct{})

	go readMessages(conn, done)

	select {
	case <-done:
		fmt.Println("Connessione chiusa dal server")
	case <-interrupt:
		fmt.Println("Interruzione ricevuta")
		conn.Close() // Chiusura semplice, non graceful
	}
}
