package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var quotePath = flag.String("quote", "", "Path to quote file")
	flag.Parse()

	if *quotePath == "" {
		fmt.Println("Usage: check-quote-format -quote <path>")
		os.Exit(1)
	}

	data, err := os.ReadFile(*quotePath)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File size: %d bytes\n", len(data))
	fmt.Println("First 100 bytes (raw):")
	fmt.Printf("%q\n", data[:min(100, len(data))])

	fmt.Println("\nFirst 100 bytes (hex):")
	fmt.Printf("%s\n", hex.EncodeToString(data[:min(100, len(data))]))

	// Check if it's JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err == nil {
		fmt.Println("\n✓ File is valid JSON")
		// Pretty print first part
		partial, _ := json.MarshalIndent(jsonData, "", "  ")
		if len(partial) > 500 {
			partial = partial[:500]
		}
		fmt.Printf("JSON content (first 500 chars):\n%s\n", partial)

		// Check if it has a quote field
		if m, ok := jsonData.(map[string]interface{}); ok {
			if quote, ok := m["quote"]; ok {
				fmt.Println("\nFound 'quote' field in JSON")
				if quoteStr, ok := quote.(string); ok {
					fmt.Printf("Quote field is a string of length %d\n", len(quoteStr))

					// Check if it's hex encoded
					if isHex(quoteStr) {
						fmt.Println("Quote appears to be hex-encoded")
						decoded, _ := hex.DecodeString(quoteStr)
						fmt.Printf("Decoded quote size: %d bytes\n", len(decoded))

						// Check header of decoded quote
						if len(decoded) >= 8 {
							version := uint16(decoded[0]) | uint16(decoded[1])<<8
							attestType := uint16(decoded[2]) | uint16(decoded[3])<<8
							teeType := uint32(decoded[4]) | uint32(decoded[5])<<8 | uint32(decoded[6])<<16 | uint32(decoded[7])<<24
							fmt.Printf("\nDecoded quote header:\n")
							fmt.Printf("  Version: %d\n", version)
							fmt.Printf("  Attestation Type: %d\n", attestType)
							fmt.Printf("  TEE Type: 0x%08x (TDX should be 0x00000081)\n", teeType)
						}
					}
				}
			}
		}
	} else {
		fmt.Println("\n✗ File is not JSON")

		// Check if it's hex-encoded text
		strData := strings.TrimSpace(string(data))
		if isHex(strData) {
			fmt.Println("✓ File appears to be hex-encoded text")
			decoded, _ := hex.DecodeString(strData)
			fmt.Printf("Decoded size: %d bytes\n", len(decoded))

			if len(decoded) >= 8 {
				version := uint16(decoded[0]) | uint16(decoded[1])<<8
				attestType := uint16(decoded[2]) | uint16(decoded[3])<<8
				teeType := uint32(decoded[4]) | uint32(decoded[5])<<8 | uint32(decoded[6])<<16 | uint32(decoded[7])<<24
				fmt.Printf("\nDecoded quote header:\n")
				fmt.Printf("  Version: %d\n", version)
				fmt.Printf("  Attestation Type: %d\n", attestType)
				fmt.Printf("  TEE Type: 0x%08x (TDX should be 0x00000081)\n", teeType)
			}
		} else {
			// Check if it's a binary quote
			if len(data) >= 8 {
				version := uint16(data[0]) | uint16(data[1])<<8
				attestType := uint16(data[2]) | uint16(data[3])<<8
				teeType := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24

				if version == 4 && teeType == 0x81 {
					fmt.Println("✓ File appears to be a binary TDX quote v4")
				} else {
					fmt.Printf("\nBinary header (if treated as binary):\n")
					fmt.Printf("  Version: %d (expected 4)\n", version)
					fmt.Printf("  Attestation Type: %d\n", attestType)
					fmt.Printf("  TEE Type: 0x%08x (expected 0x00000081 for TDX)\n", teeType)
				}
			}
		}
	}
}

func isHex(s string) bool {
	// Remove any whitespace
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")

	if len(s)%2 != 0 {
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
