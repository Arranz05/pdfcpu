package api

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// GetLastObjectID returns the highest object ID in the XRefTable.
func GetLastObjectID(xRefTable *model.XRefTable) int {
	maxID := 0
	for objID := range xRefTable.Table {
		if objID > maxID {
			maxID = objID
		}
	}
	return maxID
}

// AddObject adds a new object to the PDF and returns its object ID.
func AddObject(ctx *model.Context, object []byte, outputPath string, conf *model.Configuration) (uint32, error) {
	if ctx == nil {
		return 0, fmt.Errorf("nil PDF context")
	}

	// Determine the last object ID.
	lastXrefID := GetLastObjectID(ctx.XRefTable)
	objectID := lastXrefID + 1

	// Create PDF object entry.
	pdfObject := &types.StreamDict{
		Dict: types.Dict{
			"Length": types.Integer(len(object)),
		},
		Content:        object,
		FilterPipeline: nil,
	}
	indRef := types.IndirectRef{ObjectNumber: types.Integer(objectID)}

	ctx.XRefTable.Table[int(indRef.ObjectNumber)] = &model.XRefTableEntry{
		Object: pdfObject,
	}
	// Write the updated PDF file.
	err := WriteUpdatedPDF(ctx, outputPath, conf)
	if err != nil {
		return 0, fmt.Errorf("failed to write updated PDF: %w", err)
	}

	// Return the new object ID.
	return uint32(indRef.ObjectNumber), nil
}

// WriteUpdatedPDF writes the modified PDF context to the output file.
func WriteUpdatedPDF(ctx *model.Context, outputPath string, conf *model.Configuration) error {
	if ctx == nil {
		return fmt.Errorf("nil PDF context")
	}

	// Create or open the output file.
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			fmt.Printf("warning: failed to close output file: %v\n", err)
		}
	}()

	// Write the modified PDF to the output file.
	if err := Write(ctx, outFile, conf); err != nil {
		return fmt.Errorf("failed to write PDF context: %w", err) // TODO: The content format is a compressed object strema. Consider changing it in the future
	}

	return nil
}

// CreateSignaturePlaceholder generates a basic signature placeholder for PDF signing.
func CreateSignaturePlaceholder(signatureMaxLength int, name, location, reason, contactInfo string) []byte {
	var buf bytes.Buffer

	buf.WriteString("<<\n")
	buf.WriteString(" /Type /Sig\n")
	buf.WriteString(" /Filter /Adobe.PPKLite\n")
	buf.WriteString(" /SubFilter /adbe.pkcs7.detached\n")

	// Byte range placeholder to be replaced later.
	buf.WriteString(" /ByteRange [0 0 0 0]\n")

	// Placeholder for the actual signature content.
	buf.WriteString(" /Contents <")
	buf.Write(bytes.Repeat([]byte("0"), signatureMaxLength))
	buf.WriteString(">\n")

	// Optional metadata.
	if name != "" {
		buf.WriteString(" /Name ")
		buf.WriteString(pdfString(name))
		buf.WriteString("\n")
	}
	if location != "" {
		buf.WriteString(" /Location ")
		buf.WriteString(pdfString(location))
		buf.WriteString("\n")
	}
	if reason != "" {
		buf.WriteString(" /Reason ")
		buf.WriteString(pdfString(reason))
		buf.WriteString("\n")
	}
	if contactInfo != "" {
		buf.WriteString(" /ContactInfo ")
		buf.WriteString(pdfString(contactInfo))
		buf.WriteString("\n")
	}

	buf.WriteString(">>\n")

	return buf.Bytes()
}
func UpdateByteRange(pdfPath string, signatureOffset int, signatureLength int) error {
	// Read the entire PDF file.
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF file: %w", err)
	}

	// Calculate byte range values.
	brStart := 0
	br1 := signatureOffset            // Start of the signature
	br2 := signatureLength            // Length of the signature
	br3 := len(pdfData) - (br1 + br2) // Remaining bytes after signature

	// Format the new byte range string.
	byteRangeStr := fmt.Sprintf("/ByteRange [%d %d %d %d]", brStart, br1, br1+br2, br3)

	// Locate and replace the existing /ByteRange placeholder.
	byteRangePlaceholder := "/ByteRange [0 0 0 0]"
	updatedPDF := bytes.Replace(pdfData, []byte(byteRangePlaceholder), []byte(byteRangeStr), 1)

	if !bytes.Contains(updatedPDF, []byte(byteRangeStr)) {
		return fmt.Errorf("failed to update /ByteRange in PDF")
	}

	// Write the updated PDF back to file.
	if err := os.WriteFile(pdfPath, updatedPDF, 0644); err != nil {
		return fmt.Errorf("failed to write updated PDF file: %w", err)
	}

	return nil
}

// CreateAcroForm generates the AcroForm dictionary for the PDF catalog.
//func CreateAcroForm(ctx *model.Context) []byte {
//var buf bytes.Buffer

//buf.WriteString("<<\n")
//buf.WriteString("  /Fields [")

// Add existing signatures
//for i, sig := range existingSignatures {
//if i > 0 {
//buf.WriteString(" ")
//}
//buf.WriteString(strconv.Itoa(sig.objectId) + " 0 R")
//}

//buf.WriteString("]\n") // Close Fields array

// Signature flags
//buf.WriteString("  /SigFlags 3\n")

//buf.WriteString(">>\n") // Close AcroForm dictionary

//return buf.Bytes()
//}

// pdfString escapes and formats a string for PDF content.
func pdfString(input string) string {
	return "(" + input + ")" // Simplified escape logic.
}
