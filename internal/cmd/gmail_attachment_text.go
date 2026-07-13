package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"

	"github.com/steipete/gogcli/internal/gmailcontent"
)

// addTextContent embeds extracted text for supported attachment types (PDF,
// HTML, plain text), or a reason when extraction is not possible.
func addTextContent(payload map[string]any, path string, size int64, filename, mimeType string) {
	if size > maxInlineAttachmentBytes {
		payload["reason"] = fmt.Sprintf("attachment size %d bytes exceeds inline size limit (%d bytes); content written to path only", size, maxInlineAttachmentBytes)
		return
	}
	data, err := os.ReadFile(path) // #nosec G304 -- path is the attachment destination this command just wrote.
	if err != nil {
		payload["reason"] = fmt.Sprintf("failed to read downloaded attachment for text extraction: %v", err)
		return
	}

	// Attachment metadata can be missing (Gmail attachment IDs are unstable
	// across API responses); fall back to sniffing the content.
	mimeType = normalizeMimeType(mimeType)
	if mimeType == "" {
		mimeType = normalizeMimeType(http.DetectContentType(data))
	}

	switch {
	case isPDFAttachment(data, filename, mimeType):
		text, err := extractPDFText(data)
		if err != nil {
			payload["reason"] = fmt.Sprintf("pdf text extraction failed: %v; use --inline or --out for the raw bytes", err)
			return
		}
		payload["text"] = text
		if text == "" {
			payload["note"] = "no extractable text; the PDF may be scanned or image-only"
		}
	case isHTMLAttachment(data, mimeType):
		payload["text"] = gmailcontent.StripHTMLTags(string(data))
	case isTextAttachment(mimeType):
		payload["text"] = string(data)
	default:
		payload["reason"] = fmt.Sprintf("text extraction not supported for mimeType %q; use --inline for base64 content or --out for a local copy", mimeType)
	}
}

// normalizeMimeType lowercases and strips parameters ("; charset=...").
func normalizeMimeType(mimeType string) string {
	base, _, _ := strings.Cut(mimeType, ";")
	return strings.ToLower(strings.TrimSpace(base))
}

func isPDFAttachment(data []byte, filename, mimeType string) bool {
	return strings.EqualFold(mimeType, "application/pdf") ||
		strings.EqualFold(filepath.Ext(filename), ".pdf") ||
		bytes.HasPrefix(data, []byte("%PDF-"))
}

func isHTMLAttachment(data []byte, mimeType string) bool {
	if strings.EqualFold(mimeType, "text/html") || strings.EqualFold(mimeType, "application/xhtml+xml") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(mimeType), "text/") && gmailcontent.LooksLikeHTML(string(data))
}

func isTextAttachment(mimeType string) bool {
	lower := strings.ToLower(mimeType)
	switch {
	case strings.HasPrefix(lower, "text/"):
		return true
	case lower == "application/json", lower == "application/xml",
		lower == "application/x-yaml", lower == "application/yaml":
		return true
	}
	return false
}

// extractPDFText pulls the plain text out of a PDF. The pdf library panics on
// some malformed inputs, so recover and surface those as errors.
func extractPDFText(data []byte) (text string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	plain, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, plain); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
