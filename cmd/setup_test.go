package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/baalimago/wd-40/cmd/serve"
)

func Test_Parse(t *testing.T) {
	t.Run("it should return command if second argument specifies an existing command", func(t *testing.T) {
		want := serve.Command()
		got, err := Parse([]string{"/some/cli/path", "serve"})
		if err != nil {
			t.Fatalf(": %v", err)
		}
		if got.Describe() != want.Describe() {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})

	t.Run("it should return command if second argument specifies shortcut of specific command", func(t *testing.T) {
		want := serve.Command()
		got, err := Parse([]string{"/some/cli/path", "s"})
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
		got, gotErr := Parse([]string{"/some/cli/path", badArg})
		if got != nil {
			t.Fatalf("expected command to be nil, got: %+v", got)
		}
		if gotErr != want {
			t.Fatalf("expected: %v, got: %v", want, gotErr)
		}
	})

	t.Run("it should return NoArgsError on lack of second argument", func(t *testing.T) {
		_, gotErr := Parse([]string{"/some/cli/path"})
		if !errors.Is(gotErr, ErrNoArgs) {
			t.Fatalf("expected to get HelpfulError, got: %v", gotErr)
		}
	})
}

func TestFormatCommandDescriptions(t *testing.T) {
	// Set up mock commands
	commands = map[string]Command{
		"testCmd": serve.Command(),
	}

	// Call the function we're testing
	result := formatCommandDescriptions()

	// Check if the returned string contains the expected command descriptions
	expectedSubstring := "testCmd"
	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected formatted command descriptions to contain '%s', got '%s'", expectedSubstring, result)
	}
}
