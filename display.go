package main

import (
	"fmt"
	"strings"
	"time"
)

// Displays lightning strike information in a formatted way
func DisplayStrike(strike LightningStrike, location *NominatimResponse) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("LIGHTNING STRIKE DETECTED")
	fmt.Println(strings.Repeat("=", 80))

	// Convert timestamp to readable format
	timestamp := time.Unix(0, strike.Time)
	fmt.Printf("Time: %s\n", timestamp.Format("2006-01-02 15:04:05.000 MST"))

	// Location information
	fmt.Printf("Coordinates: %.6f, %.6f\n", strike.Lat, strike.Lon)
	fmt.Printf("Location: %s\n", FormatLocation(location))

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

	status := getStatusString(strike.Status)
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

// Formats location information from Nominatim response
func FormatLocation(location *NominatimResponse) string {
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

// Converts strike's status code to human-readable string
func getStatusString(status int) string {
	switch status {
	case 0:
		return "Very good"
	case 1:
		return "Good"
	case 2:
		return "Questionable"
	case 3:
		return "Poor"
	default:
		return "Terrible"
	}
}

// Prints the application welcome message
func PrintWelcomeMessage() {
	fmt.Println("Lightning Strike Monitor started. Press Ctrl+C to stop.")
	fmt.Println("Waiting for lightning strikes...")
	fmt.Println()
}
