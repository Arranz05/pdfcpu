package api

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// GetLastObjectID returns the highest object ID in the XRefTable.
func GetLastObjectID1(xRefTable *model.XRefTable) int {
	maxID := 0
	for objID := range xRefTable.Table {
		if objID > maxID {
			maxID = objID
		}
	}
	return maxID
}

func AddSignatureObject(ctx *model.Context, object types.Dict, outputPath string, conf *model.Configuration) (uint32, error) {
	if ctx == nil {
		return 0, fmt.Errorf("nil PDF context")
	}

	// Determine the last object ID.
	lastXrefID := GetLastObjectID1(ctx.XRefTable)
	objectID := lastXrefID + 1
	// Force pdfcpu to write as a direct object (not inside an object stream)
	ctx.Write.WriteToObjectStream = false
	ctx.Write.CurrentObjStream = nil
	ctx.Write.Increment = true

	indRef := types.IndirectRef{ObjectNumber: types.Integer(objectID)}

	ctx.XRefTable.Table[int(indRef.ObjectNumber)] = &model.XRefTableEntry{
		Object:     object,
		Compressed: false,
	}
	// Write the updated PDF file.
	err := WriteUpdatedPDF1(ctx, outputPath, conf)
	if err != nil {
		return 0, fmt.Errorf("failed to write updated PDF: %w", err)
	}

	// Return the new object ID.
	return uint32(indRef.ObjectNumber), nil
}

// WriteUpdatedPDF writes the modified PDF context to the output file.
func WriteUpdatedPDF1(ctx *model.Context, outputPath string, conf *model.Configuration) error {
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
	ctx.Write.WriteToObjectStream = false
	ctx.Write.CurrentObjStream = nil
	ctx.Write.Increment = true

	// Disable optimization to avoid object stream packing.
	//ctx.Optimize = nil

	// Write the modified PDF to the output file.
	if err := Write(ctx, outFile, conf); err != nil {
		return fmt.Errorf("failed to write PDF context: %w", err) // TODO: The content format is a compressed object strem. Consider changing it in the future
	}

	return nil
}

// CreateSignaturePlaceholder generates a basic signature placeholder dictionary for PDF signing.
func CreateSignaturePlaceholder1(signatureMaxLength int, name, location, reason, contactInfo string) types.Dict {
	placeholder := types.Dict{
		"Type":      types.Name("Sig"),
		"Filter":    types.Name("Adobe.PPKLite"),
		"SubFilter": types.Name("adbe.pkcs7.detached"),
		"ByteRange": types.Array{
			types.Integer(0),
			types.Integer(0),
			types.Integer(0),
			types.Integer(0),
		},
		// Placeholder for ByteRange
		"Contents": types.StringLiteral("<" + string(bytes.Repeat([]byte("0"), signatureMaxLength)) + ">"),
	}

	// Optional fields
	if name != "" {
		placeholder["Name"] = types.StringLiteral(name)
	}
	if location != "" {
		placeholder["Location"] = types.StringLiteral(location)
	}
	if reason != "" {
		placeholder["Reason"] = types.StringLiteral(reason)
	}
	if contactInfo != "" {
		placeholder["ContactInfo"] = types.StringLiteral(contactInfo)
	}

	return placeholder
}

func UpdateByteRange1(pdfPath string, signatureOffset int, signatureLength int) error {
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
