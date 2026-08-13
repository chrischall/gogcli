package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openclaw/gogcli/internal/ui"
)

// countProbeCapture records what the count probe actually asked Gmail for, so a
// test can assert the probe is a cheap ids-only call and not a second full
// fetch.
type countProbeCapture struct {
	fields     string
	maxResults string
	query      string
	calls      int
}

// gmailCountTestHandler serves a first page of `pageIDs` plus a count probe
// returning `totalIDs` ids. The probe is told apart from the listing by its
// `fields` selector, which is the only thing the count path sets.
func gmailCountTestHandler(t *testing.T, capture *countProbeCapture, resource string, pageIDs, totalIDs []string, probeNextPage string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		q := r.URL.Query()
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(path, "/users/me/labels"):
			_ = json.NewEncoder(w).Encode(map[string]any{"labels": []map[string]any{}})

		case strings.HasSuffix(path, "/users/me/"+resource) && strings.Contains(q.Get("fields"), "/id"):
			capture.calls++
			capture.fields = q.Get("fields")
			capture.maxResults = q.Get("maxResults")
			capture.query = q.Get("q")
			items := make([]map[string]any, 0, len(totalIDs))
			for _, id := range totalIDs {
				items = append(items, map[string]any{"id": id})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				resource: items, "nextPageToken": probeNextPage,
			})

		case strings.HasSuffix(path, "/users/me/"+resource):
			items := make([]map[string]any, 0, len(pageIDs))
			for _, id := range pageIDs {
				items = append(items, map[string]any{"id": id, "threadId": id})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{resource: items, "nextPageToken": "PAGE2"})

		case strings.Contains(path, "/users/me/threads/"), strings.Contains(path, "/users/me/messages/"):
			id := path[strings.LastIndex(path, "/")+1:]
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": id, "messages": []map[string]any{{
					"id": id, "threadId": id, "internalDate": "1760000000000",
					"payload": map[string]any{"headers": []map[string]any{
						{"name": "Subject", "value": "s"}, {"name": "From", "value": "a@b.com"},
					}},
				}},
			})

		default:
			http.NotFound(w, r)
		}
	}
}

// gmailAllCountTestHandler serves a single complete page (no nextPageToken) and
// records any count probe, which under --all must never fire.
func gmailAllCountTestHandler(t *testing.T, capture *countProbeCapture, ids []string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(path, "/users/me/labels"):
			_ = json.NewEncoder(w).Encode(map[string]any{"labels": []map[string]any{}})
		case strings.HasSuffix(path, "/users/me/threads") && strings.Contains(r.URL.Query().Get("fields"), "/id"):
			capture.calls++
			_ = json.NewEncoder(w).Encode(map[string]any{"threads": make([]map[string]any, 99)})
		case strings.HasSuffix(path, "/users/me/threads"):
			items := make([]map[string]any, 0, len(ids))
			for _, id := range ids {
				items = append(items, map[string]any{"id": id, "threadId": id})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"threads": items})
		case strings.Contains(path, "/users/me/threads/"):
			id := path[strings.LastIndex(path, "/")+1:]
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": id, "messages": []map[string]any{{
					"id": id, "threadId": id, "internalDate": "1760000000000",
					"payload": map[string]any{"headers": []map[string]any{{"name": "Subject", "value": "s"}}},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}
}

func TestGmailSearch_Count_ExactWhenSetFitsOnePage(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1", "t2", "t3"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "invoice", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	var parsed struct {
		TotalMatches        *int64 `json:"totalMatches"`
		TotalMatchesAtLeast *int64 `json:"totalMatchesAtLeast"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("decode: %v (%s)", err, result.stdout)
	}
	if parsed.TotalMatches == nil || *parsed.TotalMatches != 3 {
		t.Fatalf("totalMatches = %v, want 3", parsed.TotalMatches)
	}
	if parsed.TotalMatchesAtLeast != nil {
		t.Fatalf("an exact count must not also report a lower bound: %v", *parsed.TotalMatchesAtLeast)
	}
}

func TestGmailSearch_Count_LowerBoundWhenProbePageFills(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1", "t2"}, "MORE"))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "invoice", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	var parsed struct {
		TotalMatches        *int64 `json:"totalMatches"`
		TotalMatchesAtLeast *int64 `json:"totalMatchesAtLeast"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("decode: %v (%s)", err, result.stdout)
	}
	if parsed.TotalMatchesAtLeast == nil || *parsed.TotalMatchesAtLeast != 2 {
		t.Fatalf("totalMatchesAtLeast = %v, want 2", parsed.TotalMatchesAtLeast)
	}
	if parsed.TotalMatches != nil {
		t.Fatalf("a saturated probe must NOT be reported as an exact total: %v", *parsed.TotalMatches)
	}
}

func TestGmailSearch_Count_ProbeIsIdsOnlyAndUsesTheSameQuery(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "from:x@y.com", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}
	if capture.calls != 1 {
		t.Fatalf("count probe ran %d times, want exactly 1", capture.calls)
	}
	if capture.fields != "threads/id,nextPageToken" {
		t.Fatalf("probe fields = %q, want ids only", capture.fields)
	}
	if capture.maxResults != "500" {
		t.Fatalf("probe maxResults = %q, want 500", capture.maxResults)
	}
	if capture.query != "from:x@y.com" {
		t.Fatalf("probe query = %q, want the search's own query", capture.query)
	}
}

func TestGmailSearch_Count_AbsentWithoutTheFlag(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1", "t2"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "invoice", "--max", "1"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}
	if strings.Contains(result.stdout, "totalMatches") {
		t.Fatalf("count must be opt-in, got: %s", result.stdout)
	}
	if capture.calls != 0 {
		t.Fatalf("count probe must not run without --count, ran %d times", capture.calls)
	}
}

func TestGmailMessagesSearch_Count_Exact(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "messages",
		[]string{"m1"}, []string{"m1", "m2", "m3", "m4"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "messages", "search", "invoice", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	var parsed struct {
		TotalMatches *int64 `json:"totalMatches"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("decode: %v (%s)", err, result.stdout)
	}
	if parsed.TotalMatches == nil || *parsed.TotalMatches != 4 {
		t.Fatalf("totalMatches = %v, want 4", parsed.TotalMatches)
	}
	if capture.fields != "messages/id,nextPageToken" {
		t.Fatalf("probe fields = %q, want message ids only", capture.fields)
	}
}

func TestGmailSearch_Count_ReportsZeroForNoMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/users/me/labels") {
			_ = json.NewEncoder(w).Encode(map[string]any{"labels": []map[string]any{}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"threads": []map[string]any{}})
	}))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "zzznope", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	var parsed struct {
		TotalMatches *int64 `json:"totalMatches"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("decode: %v (%s)", err, result.stdout)
	}
	if parsed.TotalMatches == nil || *parsed.TotalMatches != 0 {
		t.Fatalf("totalMatches = %v, want 0", parsed.TotalMatches)
	}
}

func TestGmailSearch_Count_TextOutputNamesTheWholeSetOnStderr(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1", "t2", "t3"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--plain", "--account", "a@b.com", "gmail", "search", "invoice", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}
	if !strings.Contains(result.stderr, "Showing 1 of 3 matches.") {
		t.Fatalf("stderr should say the page is partial, got: %q", result.stderr)
	}
	// stdout must stay parseable.
	if strings.Contains(result.stdout, "matches.") {
		t.Fatalf("count hint leaked onto stdout: %q", result.stdout)
	}
}

func TestPrintGmailMatchCount_Wording(t *testing.T) {
	for _, tc := range []struct {
		name  string
		shown int
		count gmailMatchCount
		want  string
	}{
		{"partial exact", 3, gmailMatchCount{Value: 21, Exact: true}, "Showing 3 of 21 matches."},
		{"complete exact", 6, gmailMatchCount{Value: 6, Exact: true}, "6 matches."},
		{"lower bound", 3, gmailMatchCount{Value: 500}, "Showing 3 of at least 500 matches."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			u, err := ui.New(ui.Options{Stdout: &stdout, Stderr: &stderr, Color: "never"})
			if err != nil {
				t.Fatalf("ui.New: %v", err)
			}
			printGmailMatchCount(u, tc.shown, tc.count)
			if !strings.Contains(stderr.String(), tc.want) {
				t.Fatalf("got %q, want it to contain %q", stderr.String(), tc.want)
			}
		})
	}
}

func TestGmailSearch_Count_SkipsProbeWithAll(t *testing.T) {
	capture := &countProbeCapture{}
	// The probe would report 99 if it ran; --all must instead report what the
	// walk actually collected, so a probe result cannot masquerade as the total.
	srv := httptest.NewServer(gmailAllCountTestHandler(t, capture, []string{"t1", "t2"}))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--account", "a@b.com", "gmail", "search", "invoice", "--all", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}

	var parsed struct {
		Threads      []struct{} `json:"threads"`
		TotalMatches *int64     `json:"totalMatches"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &parsed); err != nil {
		t.Fatalf("decode: %v (%s)", err, result.stdout)
	}
	if parsed.TotalMatches == nil || *parsed.TotalMatches != 2 {
		t.Fatalf("totalMatches = %v, want 2 (what --all collected)", parsed.TotalMatches)
	}
	if capture.calls != 0 {
		t.Fatalf("--all already has the whole set; probe must not run, ran %d times", capture.calls)
	}
}

func TestGmailSearch_Count_SkipsProbeWithResultsOnly(t *testing.T) {
	capture := &countProbeCapture{}
	srv := httptest.NewServer(gmailCountTestHandler(t, capture, "threads",
		[]string{"t1"}, []string{"t1", "t2", "t3"}, ""))
	defer srv.Close()

	result := executeWithGmailTestService(t,
		[]string{"--json", "--results-only", "--account", "a@b.com", "gmail", "search", "invoice", "--max", "1", "--count"},
		newGmailServiceFromServer(t, srv))
	if result.err != nil {
		t.Fatalf("Execute: %v\nstderr=%q", result.err, result.stderr)
	}
	// --results-only unwraps to the bare array, so the count could never be
	// reported. Spending the request anyway is pure waste.
	if capture.calls != 0 {
		t.Fatalf("probe must not run under --results-only, ran %d times", capture.calls)
	}
	if !strings.Contains(result.stderr, "--count has no effect with --results-only") {
		t.Fatalf("the caller must be told, not silently ignored; stderr=%q", result.stderr)
	}
}
