package cmd

import (
	"errors"
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
		if !errors.Is(gotErr, NoArgsError) {
			t.Fatalf("expected to get HelpfulError, got: %v", gotErr)
		}
	})

}
