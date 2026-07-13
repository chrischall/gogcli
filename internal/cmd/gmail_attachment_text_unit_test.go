package cmd

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestExtractPDFTextIsolated_ExtractsViaSubprocess(t *testing.T) {
	pdf := buildMinimalPDF(t, "BT /F1 24 Tf 72 720 Td (Isolated hello) Tj ET")
	text, err := extractPDFTextIsolated(context.Background(), pdf, 30*time.Second)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if !strings.Contains(text, "Isolated hello") {
		t.Fatalf("text=%q", text)
	}
}

func TestExtractPDFTextIsolated_KillsChildOnTimeout(t *testing.T) {
	old := newPDFExtractCmd
	t.Cleanup(func() { newPDFExtractCmd = old })
	newPDFExtractCmd = func(ctx context.Context) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, os.Args[0]) // #nosec G204 -- test binary re-exec.
		cmd.Env = append(os.Environ(), "GOG_PDF_EXTRACT_CHILD=1", "GOG_TEST_PDF_EXTRACT_SLEEP=1")
		return cmd, nil
	}

	start := time.Now()
	_, err := extractPDFTextIsolated(context.Background(), []byte("%PDF-1.4"), 100*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("err=%v", err)
	}
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("hung child was not killed promptly (took %s)", elapsed)
	}
}

func TestExtractPDFTextIsolated_SurfacesChildError(t *testing.T) {
	_, err := extractPDFTextIsolated(context.Background(), []byte("%PDF-1.4\nnot a real pdf"), 30*time.Second)
	if err == nil {
		t.Fatalf("expected error for corrupt pdf")
	}
}

func TestGmailPDFExtractCmd_WritesTextToStdout(t *testing.T) {
	pdf := buildMinimalPDF(t, "BT /F1 24 Tf 72 720 Td (Child direct) Tj ET")
	var out, errBuf bytes.Buffer
	ctx := newCmdRuntimeIOContext(t, bytes.NewReader(pdf), &out, &errBuf)

	if err := (&GmailPDFExtractCmd{}).Run(ctx, &RootFlags{}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "Child direct") {
		t.Fatalf("stdout=%q", out.String())
	}
}

func TestGmailPDFExtractCmd_ErrorsOnOversizedInput(t *testing.T) {
	big := bytes.Repeat([]byte("x"), maxInlineAttachmentBytes+1)
	var out, errBuf bytes.Buffer
	ctx := newCmdRuntimeIOContext(t, bytes.NewReader(big), &out, &errBuf)

	if err := (&GmailPDFExtractCmd{}).Run(ctx, &RootFlags{}); err == nil {
		t.Fatalf("expected size error")
	}
}
