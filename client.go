package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Client orchestrates all the services
type Client struct {
	config    *Config
	ws        *WSClient
	geocoding *GeocodingService
	decoder   *LZWDecoder
	logger    *log.Logger
}

// Creates a new client with all dependencies
func NewClient(config *Config) *Client {
	geocoding := NewGeocodingService(config)
	decoder := NewLZWDecoder()
	ws := NewWSClient(config, geocoding, decoder)

	return &Client{
		config:    config,
		ws:        ws,
		geocoding: geocoding,
		decoder:   decoder,
		logger:    log.New(os.Stdout, "[Client] ", log.LstdFlags),
	}
}

// Starts the client
func (c *Client) Run() error {
	if err := c.ws.Connect(); err != nil {
		return err
	}
	defer c.ws.Close()

	// Send initial message
	initMessage := []byte(`{"a":111}`)
	if err := c.ws.SendMessage(initMessage); err != nil {
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel for interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Channel for read errors
	readErr := make(chan error, 1)

	// Start reading messages in a goroutine
	go func() {
		readErr <- c.ws.ReadMessages(ctx)
	}()

	PrintWelcomeMessage()

	// Wait for either an interrupt signal or read error
	select {
	case err := <-readErr:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("read error: %w", err)
		}
		c.logger.Println("Connection closed by server")
	case <-interrupt:
		c.logger.Println("Interrupt received, shutting down...")
		cancel() // Cancel the context to stop reading

		// Wait a bit for graceful shutdown
		select {
		case <-readErr:
		case <-time.After(5 * time.Second):
			c.logger.Println("Timeout waiting for graceful shutdown")
		}
	}

	return nil
}
