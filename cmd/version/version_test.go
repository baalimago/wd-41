package version

import (
	"context"
	"runtime/debug"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
)

func TestCommand(t *testing.T) {
	cmd := Command()

	if cmd == nil {
		t.Fatal("Expected command to be non-nil")
	}

	if cmd.Describe() != "print the version of wd-41" {
		t.Fatalf("Unexpected describe: %v", cmd.Describe())
	}

	fs := cmd.Flagset()
	if fs == nil {
		t.Fatal("Expected flagset to be non-nil")
	}

	help := cmd.Help()
	if help != "Print the version of wd-41" {
		t.Fatalf("Unexpected help output: %v", help)
	}
}

func TestRun(t *testing.T) {
	cmd := Command()
	ctx := context.Background()

	t.Run("it should print version info correctly", func(t *testing.T) {
		cmd.getVersionCmd = func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{
				Main: debug.Module{
					Version: "v1.2.3",
					Sum:     "h1:checksum",
				},
				GoVersion: "go1.18",
			}, true
		}

		// Capture output
		got := testboil.CaptureStdout(t, func(t *testing.T) {
			err := cmd.Run(ctx)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
		})

		expected := "version: v1.2.3, go version: go1.18, checksum: h1:checksum\n"
		if got != expected {
			t.Fatalf("Expected output %s, got %s", expected, got)
		}
	})
}
