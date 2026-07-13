package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func tempFilePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// newGmailAttachmentTestServicePayload serves one attachment (id a1 on
// message m1) with the given messages.get payload.
func newGmailAttachmentTestServicePayload(t *testing.T, data []byte, payload map[string]any) *gmail.Service {
	t.Helper()
	encoded := base64.URLEncoding.EncodeToString(data)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/gmail/v1/users/me/messages/m1/attachments/a1"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"data": encoded})
		case strings.Contains(r.URL.Path, "/gmail/v1/users/me/messages/m1") && !strings.Contains(r.URL.Path, "/attachments/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "m1", "payload": payload})
		default:
			http.NotFound(w, r)
		}
	})
	svc, closeServer := newGoogleTestService(t, handler, gmail.NewService)
	t.Cleanup(closeServer)
	return svc
}

// newGmailAttachmentTestService serves one attachment (id a1 on message m1)
// with the given payload metadata.
func newGmailAttachmentTestService(t *testing.T, data []byte, filename, mimeType string) *gmail.Service {
	t.Helper()
	return newGmailAttachmentTestServiceWithPayloadID(t, data, filename, mimeType, "a1")
}

// newGmailAttachmentTestServiceWithPayloadID lets the messages.get payload
// carry a different attachment ID than the one being downloaded, mimicking
// Gmail's unstable attachment IDs.
func newGmailAttachmentTestServiceWithPayloadID(t *testing.T, data []byte, filename, mimeType, payloadAttachmentID string) *gmail.Service {
	t.Helper()
	return newGmailAttachmentTestServicePayload(t, data, map[string]any{"parts": []map[string]any{{
		"filename": filename,
		"mimeType": mimeType,
		"body":     map[string]any{"attachmentId": payloadAttachmentID, "size": len(data)},
	}}})
}

// newGmailAttachmentTestServiceNoParts serves a message whose payload exposes
// no attachment parts at all.
func newGmailAttachmentTestServiceNoParts(t *testing.T, data []byte) *gmail.Service {
	t.Helper()
	return newGmailAttachmentTestServicePayload(t, data, map[string]any{})
}

func TestExecute_GmailAttachment_Inline_SmallAttachment_ReturnsBase64(t *testing.T) {
	data := []byte("hello inline attachment")
	svc := newGmailAttachmentTestService(t, data, "photo.png", "image/png")
	outPath := tempFilePath(t, "photo.png")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", outPath, "--inline",
	)

	got, ok := parsed["contentBase64"].(string)
	if !ok {
		t.Fatalf("contentBase64 missing or not string: %#v", parsed)
	}
	decoded, err := base64.StdEncoding.DecodeString(got)
	if err != nil {
		t.Fatalf("contentBase64 not decodable: %v", err)
	}
	if !bytes.Equal(decoded, data) {
		t.Fatalf("decoded=%q want=%q", decoded, data)
	}
	if parsed["mimeType"] != "image/png" {
		t.Fatalf("mimeType=%v", parsed["mimeType"])
	}
	if parsed["filename"] != "photo.png" {
		t.Fatalf("filename=%v", parsed["filename"])
	}
	if parsed["path"] != outPath {
		t.Fatalf("path=%v", parsed["path"])
	}
	if _, statErr := os.Stat(outPath); statErr != nil {
		t.Fatalf("file not written: %v", statErr)
	}
}

func TestExecute_GmailAttachment_Inline_Oversized_FallsBackToPathWithReason(t *testing.T) {
	data := bytes.Repeat([]byte("x"), maxInlineAttachmentBytes+1)
	svc := newGmailAttachmentTestService(t, data, "big.bin", "application/octet-stream")
	outPath := tempFilePath(t, "big.bin")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", outPath, "--inline",
	)

	if _, ok := parsed["contentBase64"]; ok {
		t.Fatalf("oversized attachment must not be inlined: %#v", parsed)
	}
	reason, _ := parsed["reason"].(string)
	if !strings.Contains(reason, "inline size limit") {
		t.Fatalf("reason=%q", reason)
	}
	if parsed["path"] != outPath {
		t.Fatalf("path=%v", parsed["path"])
	}
	st, statErr := os.Stat(outPath)
	if statErr != nil || st.Size() != int64(len(data)) {
		t.Fatalf("file not written: %v size=%v", statErr, st)
	}
}

func TestExecute_GmailAttachment_Default_OutputUnchanged(t *testing.T) {
	data := []byte("plain old download")
	svc := newGmailAttachmentTestService(t, data, "a.txt", "text/plain")
	outPath := tempFilePath(t, "a.txt")

	parsed := executeGmailAttachmentJSON(t, svc,
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", outPath,
	)

	for _, key := range []string{"contentBase64", "text", "reason", "mimeType", "filename"} {
		if _, ok := parsed[key]; ok {
			t.Fatalf("default output must not contain %q: %#v", key, parsed)
		}
	}
	if parsed["path"] != outPath || parsed["cached"] != false || parsed["bytes"] != float64(len(data)) {
		t.Fatalf("unexpected default output: %#v", parsed)
	}
}

func TestExecute_GmailAttachment_InlineAndText_Rejected(t *testing.T) {
	svc := newGmailAttachmentTestService(t, []byte("x"), "a.txt", "text/plain")
	result := executeWithGmailTestService(t, []string{
		"--json", "--account", "a@b.com",
		"gmail", "attachment", "m1", "a1",
		"--out", tempFilePath(t, "a.txt"), "--inline", "--text",
	}, svc)
	if result.err == nil {
		t.Fatalf("expected usage error for --inline with --text, got stdout=%q", result.stdout)
	}
}
