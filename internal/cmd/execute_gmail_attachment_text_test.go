package cmd

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// buildMinimalPDF assembles a tiny one-page PDF whose page content stream is
// given verbatim, computing the xref offsets so strict parsers accept it.
func buildMinimalPDF(t *testing.T, contentStream string) []byte {
	t.Helper()
	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contentStream), contentStream),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for i, obj := range objects {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xrefStart := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefStart)
	return buf.Bytes()
}

func TestExecute_GmailAttachment_Text_PDF_ReturnsExtractedText(t *testing.T) {
	pdfData := buildMinimalPDF(t, "BT /F1 24 Tf 72 720 Td (Hello gog attachment) Tj ET")
	svc := newGmailAttachmentTestService(t, pdfData, "report.pdf", "application/pdf")
	outPath := tempFilePath(t, "report.pdf")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", outPath, "--text",
	)

	text, ok := parsed["text"].(string)
	if !ok {
		t.Fatalf("text missing: %#v", parsed)
	}
	if !strings.Contains(text, "Hello gog attachment") {
		t.Fatalf("text=%q", text)
	}
	if parsed["mimeType"] != "application/pdf" {
		t.Fatalf("mimeType=%v", parsed["mimeType"])
	}
}

func TestExecute_GmailAttachment_Text_ImageOnlyPDF_NotesNoText(t *testing.T) {
	// A content stream with no text operators mimics a scanned/image-only PDF.
	pdfData := buildMinimalPDF(t, "q 1 0 0 1 0 0 cm Q")
	svc := newGmailAttachmentTestService(t, pdfData, "scan.pdf", "application/pdf")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "scan.pdf"), "--text",
	)

	if text, ok := parsed["text"].(string); !ok || text != "" {
		t.Fatalf("expected empty text, got %#v", parsed["text"])
	}
	note, _ := parsed["note"].(string)
	if !strings.Contains(note, "no extractable text") {
		t.Fatalf("note=%q", note)
	}
}

func TestExecute_GmailAttachment_Text_HTML_StripsTags(t *testing.T) {
	html := "<html><head><style>p{color:red}</style></head><body><p>Invoice <b>42</b> attached</p></body></html>"
	svc := newGmailAttachmentTestService(t, []byte(html), "invoice.html", "text/html")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "invoice.html"), "--text",
	)

	text, _ := parsed["text"].(string)
	if !strings.Contains(text, "Invoice 42 attached") {
		t.Fatalf("text=%q", text)
	}
	if strings.Contains(text, "<") || strings.Contains(text, "color:red") {
		t.Fatalf("tags/styles not stripped: %q", text)
	}
}

func TestExecute_GmailAttachment_Text_PlainText_ReturnsVerbatim(t *testing.T) {
	content := "line one\nline two"
	svc := newGmailAttachmentTestService(t, []byte(content), "notes.txt", "text/plain")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "notes.txt"), "--text",
	)

	if parsed["text"] != content {
		t.Fatalf("text=%q", parsed["text"])
	}
}

func TestExecute_GmailAttachment_Text_UnsupportedType_ReturnsReason(t *testing.T) {
	svc := newGmailAttachmentTestService(t, []byte{0x89, 'P', 'N', 'G'}, "pic.png", "image/png")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "pic.png"), "--text",
	)

	if _, ok := parsed["text"]; ok {
		t.Fatalf("unexpected text for image: %#v", parsed)
	}
	reason, _ := parsed["reason"].(string)
	if !strings.Contains(reason, "image/png") || !strings.Contains(reason, "--inline") {
		t.Fatalf("reason=%q", reason)
	}
	if parsed["path"] == nil {
		t.Fatalf("path missing: %#v", parsed)
	}
}

func TestExecute_GmailAttachment_Text_Oversized_ReturnsReason(t *testing.T) {
	data := bytes.Repeat([]byte("y"), maxInlineAttachmentBytes+1)
	svc := newGmailAttachmentTestService(t, data, "big.txt", "text/plain")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "big.txt"), "--text",
	)

	if _, ok := parsed["text"]; ok {
		t.Fatalf("oversized attachment must not extract: %#v", parsed)
	}
	reason, _ := parsed["reason"].(string)
	if !strings.Contains(reason, "inline size limit") {
		t.Fatalf("reason=%q", reason)
	}
}

// Gmail attachment IDs are not stable across API calls, so the metadata
// lookup may not find an exact ID match. With a single attachment on the
// message, the command should still resolve filename/mimeType from it.
func TestExecute_GmailAttachment_Text_UnstableAttachmentID_FallsBackToSingleAttachment(t *testing.T) {
	content := "sniff me"
	svc := newGmailAttachmentTestServiceWithPayloadID(t, []byte(content), "notes.txt", "text/plain", "rotated-id")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "notes.txt"), "--text",
	)

	if parsed["text"] != content {
		t.Fatalf("text=%#v", parsed["text"])
	}
	if parsed["filename"] != "notes.txt" || parsed["mimeType"] != "text/plain" {
		t.Fatalf("metadata fallback missing: %#v", parsed)
	}
}

// When no part metadata resolves at all, plain text should still extract via
// content sniffing rather than reporting an unsupported mime type.
func TestExecute_GmailAttachment_Text_NoMetadata_SniffsPlainText(t *testing.T) {
	content := "sniffed without metadata"
	svc := newGmailAttachmentTestServiceNoParts(t, []byte(content))

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "mystery.bin"), "--text",
	)

	if parsed["text"] != content {
		t.Fatalf("text=%#v (reason=%v)", parsed["text"], parsed["reason"])
	}
}

// --plain output is documented as stable TSV; multiline/tabbed extracted text
// must be escaped into a single field, not printed verbatim.
func TestExecute_GmailAttachment_Text_PlainOutput_EscapesMultiline(t *testing.T) {
	content := "line one\nline two\twith tab"
	svc := newGmailAttachmentTestService(t, []byte(content), "notes.txt", "text/plain")

	result := executeWithGmailTestService(t, []string{
		"--plain", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "notes.txt"), "--text",
	}, svc)
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	lines := strings.Split(strings.TrimRight(result.stdout, "\n"), "\n")
	if len(lines) != 6 { // path, cached, bytes, filename, mimeType, text
		t.Fatalf("expected 6 TSV rows, got %d: %q", len(lines), result.stdout)
	}
	var textLine string
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "\t")
		if !ok || strings.ContainsAny(value, "\t\n") {
			t.Fatalf("row is not a stable 2-field TSV record: %q", line)
		}
		if key == "text" {
			textLine = value
		}
	}
	unquoted, err := strconv.Unquote(textLine)
	if err != nil || unquoted != content {
		t.Fatalf("text field must round-trip via quoting: value=%q err=%v", textLine, err)
	}
}

func TestExecute_GmailAttachment_Text_CorruptPDF_ReturnsReason(t *testing.T) {
	data := []byte("%PDF-1.4\nthis is not a parseable pdf body")
	svc := newGmailAttachmentTestService(t, data, "broken.pdf", "application/pdf")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "broken.pdf"), "--text",
	)

	if _, ok := parsed["text"]; ok {
		t.Fatalf("corrupt PDF must not yield text: %#v", parsed)
	}
	reason, _ := parsed["reason"].(string)
	if !strings.Contains(reason, "pdf text extraction failed") {
		t.Fatalf("reason=%q", reason)
	}
}
