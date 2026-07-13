package cmd

import (
	"strings"
	"testing"
)

func TestParseMCPMountAnnotation(t *testing.T) {
	ann, err := parseMCPMountAnnotation("gmail_search,read")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ann.Name != "gmail_search" || ann.Risk != mcpRiskRead || ann.Service != "gmail" {
		t.Fatalf("ann = %#v", ann)
	}

	ann, err = parseMCPMountAnnotation("searchconsole_query,read,service=searchconsole")
	if err != nil {
		t.Fatalf("parse with service: %v", err)
	}
	if ann.Service != "searchconsole" {
		t.Fatalf("service = %q", ann.Service)
	}

	for _, bad := range []string{"", "gmail_search", "gmail_search,destroy", "noservicename,read", "gmail_search,read,bogus=1"} {
		if _, err := parseMCPMountAnnotation(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestParseMCPFieldAnnotation(t *testing.T) {
	ann, err := parseMCPFieldAnnotation("max,default=10,min=1,max=100")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if ann.Name != "max" || !ann.HasDefault || ann.Default != "10" || *ann.Min != 1 || *ann.Max != 100 {
		t.Fatalf("ann = %#v", ann)
	}

	ann, err = parseMCPFieldAnnotation("render,enum=RAW|USER_ENTERED,default=USER_ENTERED")
	if err != nil {
		t.Fatalf("parse enum: %v", err)
	}
	if len(ann.Enum) != 2 || ann.Enum[0] != "RAW" || ann.Enum[1] != "USER_ENTERED" {
		t.Fatalf("enum = %#v", ann.Enum)
	}

	ann, err = parseMCPFieldAnnotation("body,required,text")
	if err != nil || !ann.Required || !ann.Text {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	ann, err = parseMCPFieldAnnotation("values_json,required,json2d")
	if err != nil || !ann.JSON2D {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	ann, err = parseMCPFieldAnnotation("days,default=0,min=0,max=31,omitzero")
	if err != nil || !ann.OmitZero {
		t.Fatalf("ann = %#v, err = %v", ann, err)
	}

	for _, bad := range []string{"", "x,required,optional", "x,min=2,max=1", "x,min=abc", "x,unknown"} {
		if _, err := parseMCPFieldAnnotation(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
	if _, err := parseMCPFieldAnnotation("x,bogus"); err == nil || !strings.Contains(err.Error(), "unknown option") {
		t.Fatalf("unknown option error = %v", err)
	}
}
