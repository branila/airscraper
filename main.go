package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Lightning strike signal data from detection stations
type Signal struct {
	Alt    int     `json:"alt"`    // Altitude of the detection station in meters
	Lat    float64 `json:"lat"`    // Latitude of the detection station
	Lon    float64 `json:"lon"`    // Longitude of the detection station
	Sta    int     `json:"sta"`    // Unique ID of the detection station
	Status int     `json:"status"` // Signal quality/status from this station
	Time   int64   `json:"time"`   // Signal arrival time in microseconds (relative to strike)
}

// Lightning strike data
type LightningStrike struct {
	Alt    int      `json:"alt"`    // Altitude of the lightning strike in meters
	Delay  float64  `json:"delay"`  // Processing delay in seconds
	Lat    float64  `json:"lat"`    // Latitude of the lightning strike
	LatC   int      `json:"latc"`   // Latitude correction factor
	Lon    float64  `json:"lon"`    // Longitude of the lightning strike
	LonC   int      `json:"lonc"`   // Longitude correction factor
	MCG    int      `json:"mcg"`    // Maximum Chi-squared Goodness (localization quality)
	MDS    int      `json:"mds"`    // Maximum Distance to Stations used for triangulation
	Pol    int      `json:"pol"`    // Polarity (0 = negative, positive otherwise)
	Region int      `json:"region"` // Geographic region identifier
	Sig    []Signal `json:"sig"`    // Array of signals from detection stations
	Status int      `json:"status"` // Overall localization status (1 = good, 2 = questionable, etc.)
	Time   int64    `json:"time"`   // Strike timestamp in nanoseconds (Unix epoch)
}

// Nominatim reverse geocoding response
type NominatimResponse struct {
	PlaceID     int    `json:"place_id"`
	Licence     string `json:"licence"`
	OsmType     string `json:"osm_type"`
	OsmID       int    `json:"osm_id"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
	Address     struct {
		Road         string `json:"road"`
		Village      string `json:"village"`
		Town         string `json:"town"`
		City         string `json:"city"`
		County       string `json:"county"`
		State        string `json:"state"`
		Country      string `json:"country"`
		CountryCode  string `json:"country_code"`
		Postcode     string `json:"postcode"`
		Suburb       string `json:"suburb"`
		Municipality string `json:"municipality"`
		Province     string `json:"province"`
		Region       string `json:"region"`
	} `json:"address"`
}

// Application configuration
type Config struct {
	URL              string
	HandshakeTimeout time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	NominatimURL     string
	HTTPTimeout      time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		URL:              "wss://ws1.blitzortung.org/",
		HandshakeTimeout: 10 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		NominatimURL:     "https://nominatim.openstreetmap.org/reverse",
		HTTPTimeout:      10 * time.Second,
	}
}

type WSClient struct {
	config     *Config
	conn       *websocket.Conn
	logger     *log.Logger
	httpClient *http.Client
}

func NewWSClient(config *Config) *WSClient {
	return &WSClient{
		config: config,
		logger: log.New(os.Stdout, "[WSClient] ", log.LstdFlags),
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
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
		return fmt.Errorf("Failed to connect to WebSocket: %w", err)
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

// Performs reverse geocoding using Nominatim API
func (ws *WSClient) reverseGeocode(lat, lon float64) (*NominatimResponse, error) {
	url := fmt.Sprintf(
		"%s?format=json&lat=%f&lon=%f&zoom=18&addressdetails=1",
		ws.config.NominatimURL, lat, lon,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %w", err)
	}

	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	var result NominatimResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// Formats location information from Nominatim response
func (ws *WSClient) formatLocation(location *NominatimResponse) string {
	if location == nil {
		return "Location unknown"
	}

	var parts []string

	// Add specific location details
	if location.Address.Road != "" {
		parts = append(parts, location.Address.Road)
	}

	if location.Address.Village != "" {
		parts = append(parts, location.Address.Village)
	} else if location.Address.Town != "" {
		parts = append(parts, location.Address.Town)
	} else if location.Address.City != "" {
		parts = append(parts, location.Address.City)
	} else if location.Address.Suburb != "" {
		parts = append(parts, location.Address.Suburb)
	}

	if location.Address.County != "" {
		parts = append(parts, location.Address.County)
	}

	if location.Address.State != "" {
		parts = append(parts, location.Address.State)
	} else if location.Address.Province != "" {
		parts = append(parts, location.Address.Province)
	} else if location.Address.Region != "" {
		parts = append(parts, location.Address.Region)
	}

	if location.Address.Country != "" {
		parts = append(parts, location.Address.Country)
	}

	if len(parts) == 0 {
		return "Location unknown"
	}

	return strings.Join(parts, ", ")
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

			// Waits for a message from the WebSocket
			messageType, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure,
					websocket.CloseNormalClosure,
				) {
					return fmt.Errorf("Unexpected WebSocket error: %w", err)
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
	decodedMessage, err := ws.decodeLZW(message)
	if err != nil {
		return fmt.Errorf("failed to decode LZW: %w", err)
	}

	var strike LightningStrike
	if err := json.Unmarshal(decodedMessage, &strike); err != nil {
		ws.logger.Printf("Failed to unmarshal into LightningStrike: %v", err)
		return err
	}

	// Get location information
	location, err := ws.reverseGeocode(strike.Lat, strike.Lon)
	if err != nil {
		ws.logger.Printf("Failed to get location for strike: %v", err)
		// Continue with displaying the strike even if geocoding fails
	}

	// Add small delay to respect Nominatim rate limits
	time.Sleep(1 * time.Second)

	ws.displayStrike(strike, location)
	return nil
}

// Displays lightning strike information in a formatted way
func (ws *WSClient) displayStrike(strike LightningStrike, location *NominatimResponse) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("LIGHTNING STRIKE DETECTED")
	fmt.Println(strings.Repeat("=", 80))

	// Convert timestamp to readable format
	timestamp := time.Unix(0, strike.Time)
	fmt.Printf("Time: %s\n", timestamp.Format("2006-01-02 15:04:05.000 MST"))

	// Location information
	fmt.Printf("Coordinates: %.6f, %.6f\n", strike.Lat, strike.Lon)
	fmt.Printf("Location: %s\n", ws.formatLocation(location))

	// Strike characteristics
	fmt.Printf("Altitude: %d meters\n", strike.Alt)
	polarity := "Negative"
	if strike.Pol != 0 {
		polarity = "Positive"
	}
	fmt.Printf("Polarity: %s\n", polarity)

	// Quality metrics
	fmt.Printf("Processing delay: %.3f seconds\n", strike.Delay)
	fmt.Printf("Localization quality (MCG): %d\n", strike.MCG)
	fmt.Printf("Max distance to stations: %d meters\n", strike.MDS)

	status := "Terrible"

	switch strike.Status {
	case 0:
		status = "Very good"
	case 1:
		status = "Good"
	case 2:
		status = "Questionable"
	case 3:
		status = "Poor"
	}
	fmt.Printf("Status: %s\n", status)

	fmt.Printf("Region: %d\n", strike.Region)

	// Detection stations
	fmt.Printf("Detection stations: %d\n", len(strike.Sig))
	if len(strike.Sig) > 0 {
		fmt.Println("Station details:")
		for i, sig := range strike.Sig {
			fmt.Printf("  [%d] ID: %d, Location: %.6f, %.6f, Alt: %d m, Status: %d\n",
				i+1, sig.Sta, sig.Lat, sig.Lon, sig.Alt, sig.Status)
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
}

// Decodes LZW compressed data
func (ws *WSClient) decodeLZW(inputBytes []byte) ([]byte, error) {
	if len(inputBytes) == 0 {
		return []byte{}, nil
	}

	input := string(inputBytes)
	data := []rune(input)

	// Initialize the dictionary: codes 0-255 (ASCII characters)
	dict := make(map[int]string, 256)
	for i := range 256 {
		dict[i] = string(rune(i))
	}

	var result []byte
	prev := string(data[0])
	result = append(result, byte(data[0]))
	code := 256

	for i := 1; i < len(data); i++ {
		currCode := int(data[i])
		var entry string

		if currCode < 256 {
			entry = string(rune(currCode))
		} else if val, exists := dict[currCode]; exists {
			entry = val
		} else {
			// Special case: entry not yet in the dictionary
			if len(prev) == 0 {
				return nil, fmt.Errorf("invalid LZW data: empty previous string")
			}
			entry = prev + string(prev[0])
		}

		// Add to the decompressed string
		result = append(result, []byte(entry)...)

		// Update the dictionary
		if len(entry) > 0 {
			dict[code] = prev + string(entry[0])
			code++
		}
		prev = entry
	}

	return result, nil
}

// Run starts the WebSocket client
func (ws *WSClient) Run() error {
	if err := ws.Connect(); err != nil {
		return err
	}
	defer ws.Close()

	// Send initial message
	initMessage := []byte(`{"a":111}`)
	if err := ws.SendMessage(initMessage); err != nil {
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
		readErr <- ws.ReadMessages(ctx)
	}()

	fmt.Println("Lightning Strike Monitor started. Press Ctrl+C to stop.")
	fmt.Println("Waiting for lightning strikes...")
	fmt.Println()

	// Wait for either an interrupt signal or read error
	select {
	case err := <-readErr:
		if err != nil && err != context.Canceled {
			return fmt.Errorf("read error: %w", err)
		}
		ws.logger.Println("Connection closed by server")
	case <-interrupt:
		ws.logger.Println("Interrupt received, shutting down...")
		cancel() // Cancel the context to stop reading

		// Wait a bit for graceful shutdown
		select {
		case <-readErr:
		case <-time.After(5 * time.Second):
			ws.logger.Println("Timeout waiting for graceful shutdown")
		}
	}

	return nil
}

func main() {
	config := DefaultConfig()
	client := NewWSClient(config)

	if err := client.Run(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}
