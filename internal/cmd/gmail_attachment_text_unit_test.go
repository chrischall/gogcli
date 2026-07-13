package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunPDFExtractionWithTimeout_TimesOut(t *testing.T) {
	_, err := runPDFExtractionWithTimeout(func() (string, error) {
		time.Sleep(500 * time.Millisecond)
		return "too late", nil
	}, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("err=%v", err)
	}
}

func TestRunPDFExtractionWithTimeout_PassesThroughResult(t *testing.T) {
	text, err := runPDFExtractionWithTimeout(func() (string, error) {
		return "ok", nil
	}, time.Second)
	if err != nil || text != "ok" {
		t.Fatalf("text=%q err=%v", text, err)
	}
}

func TestRunPDFExtractionWithTimeout_PassesThroughError(t *testing.T) {
	wantErr := errors.New("boom")
	_, err := runPDFExtractionWithTimeout(func() (string, error) {
		return "", wantErr
	}, time.Second)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v", err)
	}
}
