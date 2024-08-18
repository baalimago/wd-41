package main

import (
	"context"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
	"github.com/baalimago/wd-41/cmd"
)

type MockCommand struct {
	runFunc      func(context.Context) error
	helpFunc     func() string
	describeFunc func() string
}

func (m MockCommand) Run(ctx context.Context) error {
	return m.runFunc(ctx)
}

func (m MockCommand) Help() string {
	return m.helpFunc()
}

func (m MockCommand) Describe() string {
	return m.describeFunc()
}

func (m MockCommand) Setup() error {
	return nil
}

func (m MockCommand) Flagset() *flag.FlagSet {
	return flag.NewFlagSet("test", flag.ContinueOnError)
}

func Test_printHelp_ExitCodes(t *testing.T) {
	mCmd := MockCommand{
		helpFunc:     func() string { return "Help message" },
		describeFunc: func() string { return "Describe message" },
	}
	tests := []struct {
		name     string
		command  cmd.Command
		err      error
		expected int
	}{
		{
			name:     "It should exit with code 1 on ArgNotFoundError",
			command:  mCmd,
			err:      cmd.ArgNotFoundError("test"),
			expected: 1,
		},
		{
			name:     "it should exit with code 0 on HelpfulError",
			command:  mCmd,
			err:      cmd.ErrHelpful,
			expected: 0,
		},
		{
			name:     "it should exit with code 1 on unknown errors",
			command:  mCmd,
			err:      errors.New("unknown error"),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := printHelp(tt.command, tt.err, func() {})
			if result != tt.expected {
				t.Errorf("printHelp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func Test_printHelp_output(t *testing.T) {
	t.Run("it should print cmd help on cmd.HelpfulError", func(t *testing.T) {
		want := "hello here is helpful message"
		mCmd := MockCommand{
			helpFunc: func() string { return want },
		}
		got := testboil.CaptureStdout(t, func(t *testing.T) {
			t.Helper()
			printHelp(mCmd, cmd.ErrHelpful, func() {})
		})
		// add the printline since we actually want a newline at end
		want = want + "\n"
		if got != want {
			t.Fatalf("expected: '%v', got: '%v'", want, got)
		}
	})

	t.Run("it should print error and usage on invalid argument", func(t *testing.T) {
		wantErr := "here is an error message"
		wantCode := 1
		usageHasBenePrinted := false
		mockUsagePrinter := func() {
			usageHasBenePrinted = true
		}
		gotCode := 0
		gotStdErr := testboil.CaptureStderr(t, func(t *testing.T) {
			t.Helper()
			gotCode = printHelp(MockCommand{}, cmd.ArgNotFoundError(wantErr), mockUsagePrinter)
		})

		if gotCode != wantCode {
			t.Fatalf("expected: %v, got: %v", wantCode, gotCode)
		}
		if !usageHasBenePrinted {
			t.Fatal("expected usage to have been printed")
		}
		if !strings.Contains(gotStdErr, wantErr) {
			t.Fatalf("expected stdout to contain: '%v', got out: '%v'", wantErr, gotStdErr)
		}
	})
}

func Test_Run_ExitCodes(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		argParser cmd.ArgParser
		expected  int
	}{
		{
			name: "on invalid arg, it should return exit code 1",
			args: []string{"invalid"},
			argParser: func(s []string) (cmd.Command, error) {
				return nil, errors.New("some error")
			},
			expected: 1,
		},
		{
			name: "on run error, it should return exit code 1",
			args: []string{"valid"},
			argParser: func(s []string) (cmd.Command, error) {
				return MockCommand{
					runFunc: func(ctx context.Context) error {
						return errors.New("whopsidops, error ojoj..!")
					},
				}, nil
			},
			expected: 1,
		},
		{
			name: "on success, error code 0",
			args: []string{"valid"},
			argParser: func(s []string) (cmd.Command, error) {
				return MockCommand{
					runFunc: func(ctx context.Context) error {
						return nil
					},
				}, nil
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := run(context.Background(), tt.args, tt.argParser)
			if result != tt.expected {
				t.Errorf("run() = %v, want %v", result, tt.expected)
			}
		})
	}
}
