package main

import "fmt"

// Handles LZW decompression
type LZWDecoder struct{}

// Creates a new LZW decoder
func NewLZWDecoder() *LZWDecoder {
	return &LZWDecoder{}
}

// Decodes LZW compressed data
func (d *LZWDecoder) Decode(inputBytes []byte) ([]byte, error) {
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
