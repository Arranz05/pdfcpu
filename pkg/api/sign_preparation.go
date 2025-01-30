package api

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const SignaturePlaceholderSize = 16384 // Size of the signature placeholder

// PreparePDFForSigning prepares a PDF for signing by adding a placeholder and calculating the ByteRange.
func PreparePDFForSigning(ctx *model.Context) ([]int, []byte, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("invalid PDF context")
	}

	// Access the PDF catalog
	catalogDict, err := ctx.XRefTable.Catalog()
	if err != nil {
		return nil, nil, fmt.Errorf("error accessing catalog: %v", err)
	}

	// Create or access the AcroForm dictionary
	var acroForm types.Dict
	if acroFormObj, found := catalogDict.Find("AcroForm"); found {
		acroForm, err = ctx.DereferenceDict(acroFormObj)
		if err != nil {
			return nil, nil, fmt.Errorf("error dereferencing AcroForm: %v", err)
		}
	}
	if acroForm == nil {
		acroForm = types.Dict{}
		catalogDict["AcroForm"] = acroForm
	}

	if _, found := acroForm.Find("Fields"); !found {
		acroForm["Fields"] = types.Array{}
	}

	// Define the signature field
	sigFieldDict := types.Dict{
		"Type":    types.Name("Annot"),
		"Subtype": types.Name("Widget"),
		"FT":      types.Name("Sig"),
		"T":       types.StringLiteral("Signature1"),
		"Rect":    types.Array{types.Float(100), types.Float(100), types.Float(200), types.Float(150)},
		"V":       nil,
	}
	sigFieldRef, err := ctx.IndRefForNewObject(sigFieldDict)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating signature field: %v", err)
	}

	fieldsArray, ok := acroForm["Fields"].(types.Array)
	if !ok {
		return nil, nil, fmt.Errorf("AcroForm fields are not an array")
	}
	acroForm["Fields"] = append(fieldsArray, *sigFieldRef)

	// Add the signature dictionary
	sigDict := types.Dict{
		"Type":      types.Name("Sig"),
		"Filter":    types.Name("Adobe.PPKLite"),
		"SubFilter": types.Name("adbe.pkcs7.detached"),
		"Contents":  types.HexLiteral(strings.Repeat("0", SignaturePlaceholderSize*2)),
		"ByteRange": types.Array{
			types.Integer(0), // Start of the range
			types.Integer(0), // Length before the signature
			types.Integer(0), // Start of the second range
			types.Integer(0), // Length after the signature
		},
	}
	sigFieldDict["V"] = sigDict

	// Serialize the PDF
	var buf bytes.Buffer
	ctx.Write.Writer = bufio.NewWriter(&buf)

	// Step 3: Check placeholder size before writing
	placeholderBytes := len(sigDict["Contents"].(types.HexLiteral))
	if placeholderBytes > SignaturePlaceholderSize*2 {
		return nil, nil, fmt.Errorf("signature placeholder size exceeds defined limit")
	}

	if err := pdfcpu.Write(ctx); err != nil {
		return nil, nil, fmt.Errorf("error writing PDF: %v", err)
	}

	rawPDF := buf.Bytes()

	// Step 4: Use robust method to locate the placeholder
	placeholderIndex := locatePlaceholder(rawPDF, "/Contents")
	if placeholderIndex == -1 {
		return nil, nil, fmt.Errorf("signature placeholder not found in PDF")
	}

	// Dynamically calculate ByteRange
	startOffset := placeholderIndex + len("/Contents") + 2 // Adjust for "/Contents<"
	endOffset := startOffset + SignaturePlaceholderSize
	if endOffset > len(rawPDF) {
		return nil, nil, fmt.Errorf("signature placeholder exceeds PDF size")
	}

	byteRange := []int{
		0,
		startOffset,
		endOffset,
		len(rawPDF) - endOffset,
	}

	// Extract unsigned bytes
	unsignedBytes := append(rawPDF[:startOffset], rawPDF[endOffset:]...)

	return byteRange, unsignedBytes, nil
}

// locatePlaceholder finds the index of a placeholder in the PDF.
func locatePlaceholder(pdf []byte, marker string) int {
	// Perform a case-insensitive search for the marker
	markerBytes := []byte(marker)
	for i := 0; i < len(pdf)-len(markerBytes); i++ {
		if bytes.EqualFold(pdf[i:i+len(markerBytes)], markerBytes) {
			return i
		}
	}
	return -1
}

// PreparePDFForSigningFile reads a PDF, prepares it for signing, and returns ByteRange and unsigned bytes.
func PreparePDFForSigningFile(inputPath, outputPath string) ([]int, []byte, error) {
	// Read the PDF into a context
	ctx, err := ReadContextFile(inputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading PDF file: %v", err)
	}

	// Prepare the PDF for signing
	byteRange, unsignedBytes, err := PreparePDFForSigning(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Write the updated PDF to the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(unsignedBytes); err != nil {
		return nil, nil, fmt.Errorf("error writing output PDF: %v", err)
	}
	writer.Flush()

	return byteRange, unsignedBytes, nil
}
