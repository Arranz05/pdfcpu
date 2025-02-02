package main

import (
	"fmt"
	"log"

	"github.com/pdfcpu/pdfcpu/pkg/api" // Change this to the actual import path of your `api` package.
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	inputPath := "pkg/testdata/Acroforms2.pdf"
	outputPath := "output1.pdf"

	// Load the existing PDF context.
	conf := model.NewDefaultConfiguration()
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to load PDF: %v", err)
	}

	// Create a signature placeholder object.
	signaturePlaceholder := api.CreateSignaturePlaceholder(512, "John Doe", "New York", "Approved", "johndoe@example.com")

	// Add the new object to the PDF.
	newObjectID, err := api.AddObject(ctx, signaturePlaceholder, outputPath, conf)
	if err != nil {
		log.Fatalf("Failed to add object: %v", err)
	}

	// Verify that the object was added.
	if ctx.XRefTable.Table[int(newObjectID)] == nil {
		log.Fatalf("Failed: Object %d was not added correctly", newObjectID)
	}

	fmt.Printf("Success! New object added with ID: %d\n", newObjectID)
	fmt.Printf("Updated PDF saved as: %s\n", outputPath)
}
