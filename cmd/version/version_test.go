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

	testboil.AssertStringContains(t, cmd.Describe(), "print the version of")

	fs := cmd.Flagset()
	if fs == nil {
		t.Fatal("Expected flagset to be non-nil")
	}

	testboil.AssertStringContains(t, cmd.Help(), "print the version of")
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
		testboil.AssertStringContains(t, got, expected)
	})
}
