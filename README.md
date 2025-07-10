# Lightning Strike Monitor

A real-time lightning strike monitoring application con i controcoglioni that connects to the Blitzortung network to display detailed information about lightning strikes worldwide.

## Features

- **Real-time Lightning Data**: Connects to the Blitzortung WebSocket API for live lightning strike information
- **Geographic Location**: Reverse geocoding using OpenStreetMap's Nominatim API to show strike locations
- **Detailed Strike Information**: Coordinates, polarity, quality metrics, detection stations, etc.
- **LZW Decompression**: Handles compressed data streams from the Blitzortung network

## Installation

### Prerequisites

- Go 1.19 or higher
- Internet connection for WebSocket and API access

### Dependencies

The project uses the following external library:
- `github.com/gorilla/websocket` - WebSocket client implementation

### Setup

1. Clone the repository:
```bash
git clone https://github.com/branila/airscraper.git
cd airscraper
```

3. Install dependencies:
```bash
go mod tidy
```

4. Run the application:
```bash
go run .
```

### Sample Output

```
================================================================================
LIGHTNING STRIKE DETECTED
================================================================================
Time: 2025-07-11 00:01:54.666 CEST
Coordinates: 36.312232, -78.769997
Location: Collie Jones Road, Granville County, North Carolina, United States
Altitude: 0 meters
Polarity: Negative
Processing delay: 11.400 seconds
Localization quality (MCG): 196
Max distance to stations: 8821 meters
Status: Very good
Region: 3
Detection stations: 29
Station details:
  [1] ID: 2351, Location: 39.098305, -77.522064, Alt: 108 m, Status: 4
  [2] ID: 2463, Location: 43.126045, -75.307953, Alt: 168 m, Status: 4
  [3] ID: 2005, Location: 41.614433, -72.323586, Alt: 186 m, Status: 12
  [4] ID: 1693, Location: 42.694462, -73.861351, Alt: 87 m, Status: 4
  [5] ID: 2548, Location: 42.402618, -72.515091, Alt: 109 m, Status: 12
  [6] ID: 2982, Location: 41.935860, -71.280464, Alt: 42 m, Status: 12
  [7] ID: 3012, Location: 42.363689, -71.107956, Alt: 26 m, Status: 12
  [...]
--------------------------------------------------------------------------------
```

## Architecture

The application is structured using a modular approach with clear separation of concerns:

### File Structure

```
├── main.go           # Application entry point
├── types.go          # Data structures and configuration
├── client.go         # Main client orchestrator
├── websocket.go      # WebSocket client implementation
├── geocoding.go      # Reverse geocoding service
├── lzw.go            # LZW decompression algorithm
├── display.go        # Output formatting and display
└── [...]             # Other files
```

### Components

- **Client**: Main orchestrator that coordinates all services
- **WebSocket Client**: Handles connection to Blitzortung WebSocket API
- **Geocoding Service**: Manages reverse geocoding using Nominatim API
- **LZW Decoder**: Decompresses incoming data streams
- **Display Functions**: Formats and presents lightning strike information

## Data Sources

- **Blitzortung Network**: Real-time lightning detection data
  - Website: https://www.blitzortung.org/
  - WebSocket API: `wss://ws1.blitzortung.org/`

- **OpenStreetMap Nominatim**: Reverse geocoding service
  - API: https://nominatim.openstreetmap.org/
  - Rate limit: 1 request per second (automatically handled)

## Technical Details

### Lightning Strike Data

Each lightning strike contains:
- **Coordinates**: Latitude and longitude with high precision
- **Timing**: Nanosecond precision timestamp
- **Quality Metrics**: Localization accuracy indicators
- **Detection Network**: Information from multiple detection stations
- **Physical Properties**: Altitude, polarity, and signal characteristics

### Rate Limiting

The application implements rate limiting for the Nominatim API (1 request per second) to comply with usage policies.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Disclaimer

This application is for educational and monitoring purposes only.
