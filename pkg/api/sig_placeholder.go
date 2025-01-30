package api

import (
	"bytes"
)

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

// pdfString escapes and formats a string for PDF content.
func pdfString(input string) string {
	return "(" + input + ")" // Simplified escape logic.
}
