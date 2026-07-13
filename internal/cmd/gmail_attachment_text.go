package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"

	"github.com/steipete/gogcli/internal/gmailcontent"
)

// addTextContent embeds extracted text for supported attachment types (PDF,
// HTML, plain text), or a reason when extraction is not possible.
func addTextContent(ctx context.Context, payload map[string]any, path string, size int64, filename, mimeType string) {
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
		text, err := extractPDFTextIsolated(ctx, data, pdfExtractTimeout)
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

// pdfExtractTimeout bounds how long parsing an (attacker-controlled) PDF may
// run; the input size is already capped by maxInlineAttachmentBytes.
const pdfExtractTimeout = 10 * time.Second

// maxExtractedTextBytes caps the text a PDF may expand to (compressed content
// streams can inflate well past the input size cap).
const maxExtractedTextBytes = 4 << 20

// newPDFExtractCmd builds the self-exec command for the extraction child
// (`gog gmail pdf-extract`). GOG_PDF_EXTRACT_CHILD lets the test binary
// dispatch to the child logic instead of re-running the suite.
var newPDFExtractCmd = func(ctx context.Context) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, exe, "gmail", "pdf-extract") // #nosec G204 -- re-exec of our own binary with fixed args.
	cmd.Env = append(os.Environ(), "GOG_PDF_EXTRACT_CHILD=1")
	return cmd, nil
}

// extractPDFTextIsolated parses the PDF in a separate killable process so a
// pathological attachment cannot hang or bloat the command (or a long-lived
// MCP server): on timeout the child is killed outright.
func extractPDFTextIsolated(ctx context.Context, data []byte, timeout time.Duration) (string, error) {
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd, err := newPDFExtractCmd(cctx)
	if err != nil {
		return "", err
	}
	cmd.Stdin = bytes.NewReader(data)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &cappedWriter{w: &out, remaining: maxExtractedTextBytes}
	cmd.Stderr = &cappedWriter{w: &errBuf, remaining: 4096}

	runErr := cmd.Run()
	if cctx.Err() != nil {
		return "", fmt.Errorf("timed out after %s", timeout)
	}
	if runErr != nil {
		if msg := strings.TrimSpace(errBuf.String()); msg != "" {
			return "", errors.New(msg)
		}
		return "", runErr
	}
	return out.String(), nil
}

// cappedWriter fails the write (and thereby the child) once the cap is hit.
type cappedWriter struct {
	w         io.Writer
	remaining int64
}

func (c *cappedWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > c.remaining {
		return 0, errors.New("output exceeds size limit")
	}
	c.remaining -= int64(len(p))
	return c.w.Write(p)
}

// pdfExtractChild is the child-process side: read the PDF from stdin (size
// re-checked) and return the extracted text.
func pdfExtractChild(in io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(in, maxInlineAttachmentBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxInlineAttachmentBytes {
		return "", fmt.Errorf("pdf input exceeds %d bytes", maxInlineAttachmentBytes)
	}
	return extractPDFText(data)
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
