package main

import "time"

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

type NominatimAddress struct {
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
}

// Nominatim reverse geocoding response
type NominatimResponse struct {
	PlaceID     int              `json:"place_id"`
	Licence     string           `json:"licence"`
	OsmType     string           `json:"osm_type"`
	OsmID       int              `json:"osm_id"`
	Lat         string           `json:"lat"`
	Lon         string           `json:"lon"`
	DisplayName string           `json:"display_name"`
	Address     NominatimAddress `json:"address"`
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
