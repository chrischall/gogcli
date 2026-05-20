package cmd

import (
	"strings"
	"testing"
)

// Coverage for #605: --heading-level (1-6 shortcut) and --named-style flags
// surface the Docs API paragraphStyle.namedStyleType field on `docs format`
// so existing paragraphs can be restyled into HEADING_1..HEADING_6, TITLE,
// SUBTITLE, or NORMAL_TEXT without rewriting them through the markdown path.

func TestDocsFormatFlags_HeadingLevelEmitsNamedStyleType(t *testing.T) {
	cases := []struct {
		name  string
		level int
		want  string
	}{
		{"H1", 1, "HEADING_1"},
		{"H2", 2, "HEADING_2"},
		{"H6", 6, "HEADING_6"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := (DocsFormatFlags{HeadingLevel: tt.level}).buildRequests(3, 9, "")
			if err != nil {
				t.Fatalf("buildRequests: %v", err)
			}
			if len(reqs) != 1 || reqs[0].UpdateParagraphStyle == nil {
				t.Fatalf("expected one paragraph-style request, got %#v", reqs)
			}
			pr := reqs[0].UpdateParagraphStyle
			if pr.ParagraphStyle.NamedStyleType != tt.want {
				t.Fatalf("NamedStyleType = %q, want %q", pr.ParagraphStyle.NamedStyleType, tt.want)
			}
			if !strings.Contains(pr.Fields, "namedStyleType") {
				t.Fatalf("Fields must include namedStyleType, got %q", pr.Fields)
			}
		})
	}
}

func TestDocsFormatFlags_NamedStyleAcceptsAllValidEnums(t *testing.T) {
	for _, ns := range []string{"NORMAL_TEXT", "TITLE", "SUBTITLE", "HEADING_1", "HEADING_2", "HEADING_3", "HEADING_4", "HEADING_5", "HEADING_6"} {
		t.Run(ns, func(t *testing.T) {
			reqs, err := (DocsFormatFlags{NamedStyle: ns}).buildRequests(1, 2, "")
			if err != nil {
				t.Fatalf("buildRequests(%q): %v", ns, err)
			}
			if got := reqs[0].UpdateParagraphStyle.ParagraphStyle.NamedStyleType; got != ns {
				t.Fatalf("NamedStyleType = %q, want %q", got, ns)
			}
		})
	}
}

func TestDocsFormatFlags_NamedStyleIsCaseInsensitive(t *testing.T) {
	reqs, err := (DocsFormatFlags{NamedStyle: "heading_3"}).buildRequests(1, 2, "")
	if err != nil {
		t.Fatalf("buildRequests: %v", err)
	}
	if got := reqs[0].UpdateParagraphStyle.ParagraphStyle.NamedStyleType; got != "HEADING_3" {
		t.Fatalf("NamedStyleType = %q, want HEADING_3", got)
	}
}

func TestDocsFormatFlags_HeadingLevelAndNamedStyleMutuallyExclusive(t *testing.T) {
	_, err := (DocsFormatFlags{HeadingLevel: 1, NamedStyle: "TITLE"}).buildRequests(1, 2, "")
	if err == nil || !strings.Contains(err.Error(), "cannot be combined") {
		t.Fatalf("expected mutual-exclusion error, got %v", err)
	}
}

func TestDocsFormatFlags_HeadingLevelOutOfRangeRejected(t *testing.T) {
	for _, lvl := range []int{-1, 0, 7, 99} {
		if lvl == 0 {
			// 0 means "no heading level" — not an error on its own.
			continue
		}
		_, err := (DocsFormatFlags{HeadingLevel: lvl}).buildRequests(1, 2, "")
		if err == nil || !strings.Contains(err.Error(), "--heading-level must be between 1 and 6") {
			t.Fatalf("level %d: expected range error, got %v", lvl, err)
		}
	}
}

func TestDocsFormatFlags_UnknownNamedStyleRejected(t *testing.T) {
	_, err := (DocsFormatFlags{NamedStyle: "BANNER"}).buildRequests(1, 2, "")
	if err == nil || !strings.Contains(err.Error(), "--named-style must be one of") {
		t.Fatalf("expected named-style enum error, got %v", err)
	}
}

func TestDocsFormatFlags_HeadingComposesWithAlignment(t *testing.T) {
	reqs, err := (DocsFormatFlags{HeadingLevel: 1, Alignment: "center"}).buildRequests(3, 9, "t.tab")
	if err != nil {
		t.Fatalf("buildRequests: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected one paragraph request combining alignment + heading, got %d", len(reqs))
	}
	pr := reqs[0].UpdateParagraphStyle
	if pr.ParagraphStyle.NamedStyleType != "HEADING_1" || pr.ParagraphStyle.Alignment != "CENTER" {
		t.Fatalf("unexpected style: %#v", pr.ParagraphStyle)
	}
	if !strings.Contains(pr.Fields, "namedStyleType") || !strings.Contains(pr.Fields, "alignment") {
		t.Fatalf("Fields missing combined attrs: %q", pr.Fields)
	}
	if pr.Range.TabId != "t.tab" {
		t.Fatalf("range lost tab: %#v", pr.Range)
	}
}

func TestDocsFormatFlags_AnyDetectsHeadingFlags(t *testing.T) {
	if !(DocsFormatFlags{HeadingLevel: 2}).any() {
		t.Fatalf("any() should be true when HeadingLevel is set")
	}
	if !(DocsFormatFlags{NamedStyle: "TITLE"}).any() {
		t.Fatalf("any() should be true when NamedStyle is set")
	}
	if (DocsFormatFlags{}).any() {
		t.Fatalf("any() should be false when nothing is set")
	}
}
