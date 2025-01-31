package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api" // Replace with your actual module name
)

func main() {
	// Define the input and output PDF file paths
	inputPath := filepath.Join("pkg/testdata/Acroforms2.pdf") // Replace with your input PDF file path
	outputPath := filepath.Join("output.pdf")                 // Replace with your output PDF file path

	// Call the PreparePDFForSigningFile function
	byteRange, unsignedBytes, err := api.PreparePDFForSigningFile(inputPath, outputPath)
	if err != nil {
		log.Fatalf("Error: Failed to prepare PDF for signing: %v", err)
	}

	// Print the ByteRange and unsigned bytes for verification
	fmt.Printf("ByteRange: %v\n", byteRange)
	fmt.Printf("Unsigned Bytes (hex, first 25 bytes): %x\n", unsignedBytes[:25]) // Limit to 25 bytes for display

	// Notify the user of the output file
	fmt.Printf("Prepared PDF saved to: %s\n", outputPath)
}
