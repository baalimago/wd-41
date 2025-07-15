package cmd

import (
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
	"golang.org/x/net/context"
)

type mockCommand struct {
	runFunc      func(context.Context) error
	helpFunc     func() string
	describeFunc func() string
	setupFunc    func() error
	flagSet      *flag.FlagSet
}

func (m *mockCommand) Run(ctx context.Context) error {
	return m.runFunc(ctx)
}

func (m *mockCommand) Help() string {
	return m.helpFunc()
}

func (m *mockCommand) Describe() string {
	return m.describeFunc()
}

func (m *mockCommand) Setup(ctx context.Context) error {
	return m.setupFunc()
}

func (m *mockCommand) Flagset() *flag.FlagSet {
	if m.flagSet != nil {
		return m.flagSet
	}
	return flag.NewFlagSet("test", flag.ContinueOnError)
}

func Test_Parse(t *testing.T) {
	t.Run("it should return command if second argument specifies an existing command", func(t *testing.T) {
		want := &mockCommand{
			describeFunc: func() string { return "serve" },
		}
		got, err := parse([]string{"/some/cli/path", "serve"}, map[string]Command{"serve": want})
		if err != nil {
			t.Fatalf(": %v", err)
		}
		if got.Describe() != want.Describe() {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should return command if second argument specifies shortcut of specific command", func(t *testing.T) {
		want := &mockCommand{
			describeFunc: func() string { return "serve" },
		}
		got, err := parse([]string{"/some/cli/path", "serve"}, map[string]Command{"serve": want})
		if err != nil {
			t.Fatalf(": %v", err)
		}
		if got.Describe() != want.Describe() {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should return error if command doesnt exist", func(t *testing.T) {
		badArg := "blheruh"
		want := ArgNotFoundError(badArg)
		got, gotErr := parse([]string{"/some/cli/path", badArg}, map[string]Command{})
		if got != nil {
			t.Fatalf("expected command to be nil, got: %+v", got)
		}
		if gotErr != want {
			t.Fatalf("expected: %v, got: %v", want, gotErr)
		}
	})

	t.Run("it should return NoArgsError on lack of second argument", func(t *testing.T) {
		_, gotErr := parse([]string{"/some/cli/path"}, map[string]Command{})
		if !errors.Is(gotErr, ErrNoArgs) {
			t.Fatalf("expected to get HelpfulError, got: %v", gotErr)
		}
	})
}

func Test_formatCommandDescriptions(t *testing.T) {
	want := &mockCommand{
		describeFunc: func() string { return "testCmd" },
	}

	// Set up mock commands
	commands := map[string]Command{
		"testCmd": want,
	}

	// Call the function we're testing
	result := formatCommandDescriptions(commands)

	// Check if the returned string contains the expected command descriptions
	expectedSubstring := "testCmd"
	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected formatted command descriptions to contain '%s', got '%s'", expectedSubstring, result)
	}
}

func Test_printHelp_ExitCodes(t *testing.T) {
	mCmd := mockCommand{
		helpFunc:     func() string { return "Help message" },
		describeFunc: func() string { return "Describe message" },
	}
	tests := []struct {
		name     string
		command  Command
		err      error
		expected int
	}{
		{
			name:     "It should exit with code 1 on ArgNotFoundError",
			command:  &mCmd,
			err:      ArgNotFoundError("test"),
			expected: 1,
		},
		{
			name:     "it should exit with code 0 on HelpfulError",
			command:  &mCmd,
			err:      flag.ErrHelp,
			expected: 0,
		},
		{
			name:     "it should exit with code 1 on unknown errors",
			command:  &mCmd,
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
	t.Run("it should print cmd help on HelpfulError", func(t *testing.T) {
		want := "hello here is helpful message"
		mCmd := &mockCommand{
			helpFunc: func() string { return want },
		}
		got := testboil.CaptureStdout(t, func(t *testing.T) {
			t.Helper()
			printHelp(mCmd, flag.ErrHelp, func() {})
		})
		// add the printline since we actually want a newline at end
		want = want + "\n"
		if strings.Contains(got, want) {
			t.Fatalf("expected: '%v', got: '%v'", want, got)
		}
	})

	t.Run("it should print error and usage on invalid argument", func(t *testing.T) {
		wantErr := "here is an error message"
		wantCode := 1
		usageHasBeenPrinted := false
		mockUsagePrinter := func() {
			usageHasBeenPrinted = true
		}
		gotCode := 0
		gotStdErr := testboil.CaptureStderr(t, func(t *testing.T) {
			t.Helper()
			gotCode = printHelp(&mockCommand{}, ArgNotFoundError(wantErr), mockUsagePrinter)
		})

		if gotCode != wantCode {
			t.Fatalf("expected: %v, got: %v", wantCode, gotCode)
		}
		if !usageHasBeenPrinted {
			t.Fatal("expected usage to have been printed")
		}
		if !strings.Contains(gotStdErr, wantErr) {
			t.Fatalf("expected stdout to contain: '%v', got out: '%v'", wantErr, gotStdErr)
		}
	})
}

func Test_Run(t *testing.T) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.Int("intVal", 0, "Set some int here not a string thanks")

	commands := map[string]Command{
		"withSetupError": &mockCommand{
			describeFunc: func() string { return "hello description" },
			setupFunc: func() error {
				return errors.New("this no work")
			},
		},
		"withError": &mockCommand{
			runFunc:      func(ctx context.Context) error { return errors.New("here error") },
			describeFunc: func() string { return "hello description" },
			setupFunc: func() error {
				return nil
			},
		},
		"valid": &mockCommand{
			runFunc:      func(ctx context.Context) error { return nil },
			describeFunc: func() string { return "hello description" },
			setupFunc: func() error {
				return nil
			},
			flagSet: flags,
		},
	}
	tests := []struct {
		name                string
		args                []string
		expectedCode        int
		expectedOutContains string
		expectedErrContains string
	}{
		{
			name:         "on invalid arg, it should return exit code 1",
			args:         []string{"bin invalid"},
			expectedCode: 1,
		},
		{
			name:                "on run error, it should return exit code 1",
			args:                []string{"bin", "withError"},
			expectedCode:        1,
			expectedErrContains: "failed to run",
		},
		{
			name:         "on success, error code 0",
			args:         []string{"bin", "valid"},
			expectedCode: 0,
		},
		{
			name:                "on bad flag, flagset should error",
			args:                []string{"bin", "-intVal=thisisstring", "valid"},
			expectedCode:        1,
			expectedErrContains: "failed to parse flagset",
		},
		{
			name:                "on bad flag, flagset should error, flag after arg",
			args:                []string{"bin", "valid", "-intVal=thisisstring"},
			expectedCode:        1,
			expectedErrContains: "failed to parse flagset",
		},
		{
			name:                "on setup error, error code 1",
			args:                []string{"bin", "withSetupError"},
			expectedCode:        1,
			expectedErrContains: "failed to setup command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOut string
			gotErr := testboil.CaptureStderr(t, func(t *testing.T) {
				gotOut = testboil.CaptureStdout(t, func(t *testing.T) {
					result := Run(context.Background(), tt.args,
						commands,
						"usage")
					if result != tt.expectedCode {
						t.Errorf("run() = %v, want %v", result, tt.expectedCode)
					}
				})
			})

			if tt.expectedOutContains != "" {
				if !strings.Contains(gotOut, tt.expectedOutContains) {
					t.Errorf("expected: '%v' to contain: '%v'", gotOut, tt.expectedOutContains)
				}
			}

			if tt.expectedErrContains != "" {
				if !strings.Contains(gotErr, tt.expectedErrContains) {
					t.Errorf("expected: '%v' to contain: '%v'", gotOut, tt.expectedErrContains)
				}
			}
		})
	}
}
