package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// mcpSuites maps a suite name to the service areas it bundles. Suites are
// curated, cross-service collections selected with `gog mcp --suite <name>`.
// They are orthogonal to per-service gating: every service remains independently
// selectable via --allow-tool regardless of suite membership, and --suite
// composes with --allow-tool (intersection) and --allow-write.
func mcpSuites() map[string][]string {
	return map[string][]string{
		"workspace": {"gmail", "calendar", "drive", "docs", "sheets", "slides", "contacts", "people", "tasks", "chat", "keep", "meet", "forms"},
		"developer": {"appscript", "api"},
		"admin":     {"admin", "groups"},
		"education": {"classroom"},
		"media":     {"photos", "youtube"},
		"insights":  {"searchconsole"},
	}
}

// mcpSuiteNames returns the available suite names in stable order.
func mcpSuiteNames() []string {
	suites := mcpSuites()
	names := make([]string, 0, len(suites))
	for name := range suites {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// mcpResolveSuites validates the requested suite names and returns the union of
// their services as a set. An unknown suite name is an error so misconfiguration
// fails loudly at server startup rather than silently exposing nothing.
func mcpResolveSuites(requested []string) (map[string]bool, error) {
	suites := mcpSuites()
	services := map[string]bool{}
	for _, raw := range requested {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		members, ok := suites[name]
		if !ok {
			return nil, fmt.Errorf("unknown suite %q (available: %s)", name, strings.Join(mcpSuiteNames(), ", "))
		}
		for _, svc := range members {
			services[svc] = true
		}
	}
	return services, nil
}

// mcpSuiteServiceSet resolves suites into a service set, ignoring unknown names.
// Callers that need validation should use mcpResolveSuites; mcpEnabledTools uses
// this after Run has already validated the names.
func mcpSuiteServiceSet(requested []string) map[string]bool {
	services, _ := mcpResolveSuites(requested)
	if services == nil {
		// On a validation error mcpResolveSuites returns a nil map; rebuild from
		// the known suites so a partially-valid request still filters sensibly.
		suites := mcpSuites()
		services = map[string]bool{}
		for _, raw := range requested {
			if members, ok := suites[strings.TrimSpace(raw)]; ok {
				for _, svc := range members {
					services[svc] = true
				}
			}
		}
	}
	return services
}

// mcpPrintSuites writes the suite-to-services map as JSON.
func mcpPrintSuites(output io.Writer) error {
	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{"suites": mcpSuites()})
}
