package main

import (
	"fmt"
	"log"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	// Define input and output PDF file paths
	inputPath := "output1.pdf"
	outputPath := "output2.pdf"

	// Load the existing PDF context
	conf := model.NewDefaultConfiguration()
	conf.WriteObjectStream = false // ✅ Prevent writing to object streams
	conf.WriteXRefStream = false   // ✅ Prevent writing XRef streams

	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to load PDF: %v", err)
	}

	// ✅ Ensure no existing object streams are used
	if ctx.Read.UsingObjectStreams {
		fmt.Println("Disabling existing object streams...")
		ctx.Read.UsingObjectStreams = false
		ctx.Read.ObjectStreams = nil
	}

	// ✅ Ensure new objects are written as regular indirect objects
	ctx.Write.WriteToObjectStream = false
	ctx.Write.CurrentObjStream = nil
	ctx.Write.Increment = true

	// Create the signature placeholder
	signaturePlaceholder := api.CreateSignaturePlaceholder1(16384, "John Doe", "New York", "Approval", "johndoe@example.com")

	// Add the signature placeholder to the PDF
	num, err := api.AddSignatureObject(ctx, signaturePlaceholder, outputPath, conf)
	if err != nil {
		fmt.Printf("Error adding signature placeholder: %v\n", err)
		return
	}

	fmt.Println("Signature placeholder added successfully. Output saved to:", outputPath)
	fmt.Printf("Signature placeholder number: %d\n", num)
}
