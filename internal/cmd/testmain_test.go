package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/steipete/gogcli/internal/tzembed" // Embed IANA timezone database for Windows test support
)

func TestMain(m *testing.M) {
	// Isolated PDF extraction re-execs the current binary; under `go test`
	// that is this test binary, so dispatch to the child logic instead of
	// recursively running the suite.
	if os.Getenv("GOG_PDF_EXTRACT_CHILD") == "1" {
		if os.Getenv("GOG_TEST_PDF_EXTRACT_SLEEP") == "1" {
			time.Sleep(time.Hour)
		}
		text, err := pdfExtractChild(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(text)
		os.Exit(0)
	}

	contactsSearchWarmupDelay = 0

	root, err := os.MkdirTemp("", "gogcli-tests-*")
	if err != nil {
		panic(err)
	}

	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")

	home := filepath.Join(root, "home")
	xdg := filepath.Join(root, "xdg")
	_ = os.MkdirAll(home, 0o755)
	_ = os.MkdirAll(xdg, 0o755)
	_ = os.Setenv("HOME", home)
	_ = os.Setenv("XDG_CONFIG_HOME", xdg)

	code := m.Run()

	if oldHome == "" {
		_ = os.Unsetenv("HOME")
	} else {
		_ = os.Setenv("HOME", oldHome)
	}
	if oldXDG == "" {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
	} else {
		_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}
	_ = os.RemoveAll(root)
	os.Exit(code)
}
