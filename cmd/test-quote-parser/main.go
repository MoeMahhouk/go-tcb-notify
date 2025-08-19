package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/MoeMahhouk/go-tcb-notify/pkg/tdx"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		quotePath = flag.String("quote", "", "Path to quote file (binary or hex)")
		hexInput  = flag.Bool("hex", false, "Input is hex encoded")
		verbose   = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if *quotePath == "" {
		fmt.Println("Usage: test-quote-parser -quote <path> [-hex] [-v]")
		os.Exit(1)
	}

	// Read quote data
	data, err := os.ReadFile(*quotePath)
	if err != nil {
		logrus.Fatal("Failed to read quote file:", err)
	}

	// Decode hex if needed
	var quoteData []byte
	if *hexInput {
		quoteData, err = hex.DecodeString(string(data))
		if err != nil {
			logrus.Fatal("Failed to decode hex:", err)
		}
	} else {
		quoteData = data
	}

	// Parse the quote
	parser := tdx.NewQuoteParser()
	parsedQuote, err := parser.ParseQuote(quoteData)
	if err != nil {
		logrus.Fatal("Failed to parse quote:", err)
	}

	// Display results
	fmt.Println("=== TDX Quote Parse Results ===")
	fmt.Printf("FMSPC: %s\n", parsedQuote.FMSPC)

	// Get and display quote info
	info := parser.GetQuoteInfo(parsedQuote)

	if version, ok := info["version"]; ok {
		fmt.Printf("Quote Version: %v\n", version)
	}
	if teeType, ok := info["teeType"]; ok {
		// Fix the display - it's already a hex string, don't double-format it
		fmt.Printf("TEE Type: %v\n", teeType)
	}
	if attestKeyType, ok := info["attestationKeyType"]; ok {
		fmt.Printf("Attestation Key Type: %v\n", attestKeyType)
	}

	fmt.Println("\n=== TCB Components ===")
	fmt.Println("SGX TCB Components:")
	for i, comp := range parsedQuote.TCBComponents.SGX {
		if comp != 0 {
			fmt.Printf("  Component[%d]: %d\n", i, comp)
		}
	}

	fmt.Println("TDX TCB Components:")
	for i, comp := range parsedQuote.TCBComponents.TDX {
		if comp != 0 {
			fmt.Printf("  Component[%d]: %d\n", i, comp)
		}
	}

	fmt.Printf("PCE SVN: %d\n", parsedQuote.TCBComponents.PCESVN)

	// Optionally verify signature
	fmt.Println("\n=== Signature Verification ===")
	err = parser.VerifyQuoteSignature(quoteData, false)
	if err != nil {
		fmt.Printf("Signature verification failed: %v\n", err)
	} else {
		fmt.Println("Signature verification passed")
	}

	// Output JSON for further processing
	if *verbose {
		tcbJSON, _ := json.MarshalIndent(parsedQuote.TCBComponents, "", "  ")
		fmt.Println("\n=== TCB Components JSON ===")
		fmt.Println(string(tcbJSON))
	}
}

// Example usage:
// go run cmd/test-quote-parser/main.go -quote sample_quote.bin -v
// go run cmd/test-quote-parser/main.go -quote sample_quote.hex -hex
