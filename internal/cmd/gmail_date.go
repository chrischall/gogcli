package cmd

import (
	"net/mail"
	"strings"
	"time"
)

const (
	// listDateLayout renders message-list and thread dates, e.g. "2026-06-17 00:29".
	listDateLayout = "2006-01-02 15:04"
	// quoteDateLayout renders dates in the Gmail-style human form used for reply
	// attributions and forwarded-message headers, e.g. "Wed, Jun 17, 2026 at 12:29 AM".
	quoteDateLayout = "Mon, Jan 2, 2006 at 3:04 PM"
)

// formatMailDateInLocation parses an RFC 2822 Date header and renders it with
// layout, converted into loc (a nil loc falls back to time.Local). Empty input
// and dates that cannot be parsed are returned trimmed and otherwise unchanged,
// so an unexpected Date value never breaks the output. A parseable date's
// weekday is recomputed from the instant in loc, which also corrects a wrong
// weekday in the source header.
func formatMailDateInLocation(raw string, loc *time.Location, layout string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	if t, err := mailParseDate(raw); err == nil {
		return t.In(loc).Format(layout)
	}
	return raw
}

// formatGmailDateInLocation renders a Date header for message and thread listings.
func formatGmailDateInLocation(raw string, loc *time.Location) string {
	return formatMailDateInLocation(raw, loc, listDateLayout)
}

// formatGmailDateISO renders a message's internalDate as RFC3339 in loc.
//
// Two things make this the value a machine consumer should read:
//
//   - It carries an explicit offset. listDateLayout ("2006-01-02 15:04") does
//     not, so a caller that sees "2026-07-28 03:36" has to already know which
//     zone gog was configured for. Get that wrong and the timestamp can land on
//     the wrong calendar DAY, which is what makes it worse than merely
//     imprecise.
//   - internalDate is the Gmail API's own receipt timestamp (epoch
//     milliseconds, UTC). The Date header is written by the sending client and
//     varies in format, offset, and honesty; internalDate does not.
//
// Returns "" when internalDate is absent, so the field stays omitempty rather
// than claiming the Unix epoch.
func formatGmailDateISO(internalDateMillis int64, loc *time.Location) string {
	if internalDateMillis <= 0 {
		return ""
	}
	if loc == nil {
		loc = time.Local
	}
	return time.UnixMilli(internalDateMillis).In(loc).Format(time.RFC3339)
}

// formatQuoteDate renders an original message's Date header in the Gmail-style
// human form used for quoted replies and forwarded-message headers, converted
// into loc (matching how Gmail shows the attribution in the reader's timezone).
func formatQuoteDate(raw string, loc *time.Location) string {
	return formatMailDateInLocation(raw, loc, quoteDateLayout)
}

func mailParseDate(s string) (time.Time, error) {
	// net/mail has the most compatible Date parser, but we keep this isolated for easier tests/mocks later.
	return mail.ParseDate(s)
}
