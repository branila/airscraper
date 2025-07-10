package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// Handles WebSocket communication
type WSClient struct {
	config    *Config
	conn      *websocket.Conn
	logger    *log.Logger
	geocoding *GeocodingService
	decoder   *LZWDecoder
}

// Creates a new WebSocket client
func NewWSClient(config *Config, geocoding *GeocodingService, decoder *LZWDecoder) *WSClient {
	return &WSClient{
		config:    config,
		logger:    log.New(os.Stdout, "[WSClient] ", log.LstdFlags),
		geocoding: geocoding,
		decoder:   decoder,
	}
}

// Establishes a WebSocket connection
func (ws *WSClient) Connect() error {
	ws.logger.Printf("Connecting to %s...", ws.config.URL)

	dialer := &websocket.Dialer{
		HandshakeTimeout: ws.config.HandshakeTimeout,
	}

	conn, _, err := dialer.Dial(ws.config.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.conn = conn
	ws.logger.Printf("Connection established with %s", ws.config.URL)

	return nil
}

// Closes the WebSocket connection
func (ws *WSClient) Close() error {
	if ws.conn == nil {
		return nil
	}

	// Send close message
	err := ws.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		ws.logger.Printf("Error sending close message: %v", err)
	}

	// Close the connection
	return ws.conn.Close()
}

// Sends a message to the WebSocket
func (ws *WSClient) SendMessage(message []byte) error {
	if ws.conn == nil {
		return fmt.Errorf("connection not established")
	}

	ws.conn.SetWriteDeadline(time.Now().Add(ws.config.WriteTimeout))
	err := ws.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	ws.logger.Println("Message sent successfully")
	return nil
}

// Reads messages from the WebSocket connection
func (ws *WSClient) ReadMessages(ctx context.Context) error {
	if ws.conn == nil {
		return fmt.Errorf("connection not established")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ws.conn.SetReadDeadline(time.Now().Add(ws.config.ReadTimeout))

			// Wait for a message from the WebSocket
			messageType, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure,
				) {
					return fmt.Errorf("unexpected WebSocket error: %w", err)
				}

				return err
			}

			if messageType == websocket.CloseMessage {
				return nil
			}

			if err := ws.processMessage(message); err != nil {
				ws.logger.Printf("Error processing message: %v", err)
				continue
			}
		}
	}
}

// Processes a received message
func (ws *WSClient) processMessage(message []byte) error {
	decodedMessage, err := ws.decoder.Decode(message)
	if err != nil {
		return fmt.Errorf("failed to decode LZW: %w", err)
	}

	var strike LightningStrike
	if err := json.Unmarshal(decodedMessage, &strike); err != nil {
		ws.logger.Printf("Failed to unmarshal into LightningStrike: %v", err)
		return err
	}

	// Get location information
	location, err := ws.geocoding.ReverseGeocodeWithRateLimit(strike.Lat, strike.Lon)
	if err != nil {
		ws.logger.Printf("Failed to get location for strike: %v", err)
		// Continue with displaying the strike even if geocoding fails!
	}

	DisplayStrike(strike, location)

	return nil
}
