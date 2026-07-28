package cmd

import (
	"testing"
	"time"
)

// formatGmailDateISO exists because listDateLayout ("2006-01-02 15:04") carries
// no offset: a caller reading "2026-07-28 03:36" has to already know which zone
// gog was configured for, and getting that wrong can move the timestamp onto the
// wrong calendar DAY.
func TestFormatGmailDateISO(t *testing.T) {
	eastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tzdata unavailable: %v", err)
	}

	// 2026-07-28T03:36:00Z is 11:36 PM Eastern on July 27 — the previous day.
	// This is the exact shape of the reported failure.
	julyUTC := time.Date(2026, 7, 28, 3, 36, 0, 0, time.UTC).UnixMilli()
	got := formatGmailDateISO(julyUTC, eastern)
	if want := "2026-07-27T23:36:00-04:00"; got != want {
		t.Fatalf("July: got %q, want %q", got, want)
	}
	if legacy := formatGmailDateInLocation("Tue, 28 Jul 2026 03:36:00 +0000", eastern); legacy != "2026-07-27 23:36" {
		t.Fatalf("legacy layout drifted: %q", legacy)
	}

	// DST comes from the IANA database, so January is -05:00, not -04:00.
	jan := time.Date(2026, 1, 15, 17, 0, 0, 0, time.UTC).UnixMilli()
	if got := formatGmailDateISO(jan, eastern); got != "2026-01-15T12:00:00-05:00" {
		t.Fatalf("January: got %q, want 2026-01-15T12:00:00-05:00", got)
	}
}

// An absent internalDate must yield "" so the field stays omitempty rather than
// claiming the Unix epoch.
func TestFormatGmailDateISO_MissingInternalDate(t *testing.T) {
	for _, ms := range []int64{0, -1} {
		if got := formatGmailDateISO(ms, time.UTC); got != "" {
			t.Fatalf("internalDate %d: got %q, want empty", ms, got)
		}
	}
}

// A nil location falls back to time.Local rather than panicking, matching
// formatMailDateInLocation.
func TestFormatGmailDateISO_NilLocation(t *testing.T) {
	got := formatGmailDateISO(time.Date(2026, 7, 28, 3, 36, 0, 0, time.UTC).UnixMilli(), nil)
	if got == "" {
		t.Fatal("nil location produced no output")
	}
	if _, err := time.Parse(time.RFC3339, got); err != nil {
		t.Fatalf("nil location produced unparseable RFC3339 %q: %v", got, err)
	}
}

// Whatever the zone, the value round-trips to the same instant — the offset is
// carried, not dropped.
func TestFormatGmailDateISO_RoundTripsToSameInstant(t *testing.T) {
	ms := time.Date(2026, 7, 28, 3, 36, 0, 0, time.UTC).UnixMilli()
	for _, name := range []string{"America/New_York", "UTC", "Asia/Tokyo", "Asia/Kolkata"} {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Skipf("tzdata unavailable for %s: %v", name, err)
		}
		got := formatGmailDateISO(ms, loc)
		parsed, parseErr := time.Parse(time.RFC3339, got)
		if parseErr != nil {
			t.Fatalf("%s: unparseable %q: %v", name, got, parseErr)
		}
		if parsed.UnixMilli() != ms {
			t.Fatalf("%s: %q round-tripped to %d, want %d", name, got, parsed.UnixMilli(), ms)
		}
	}
}
