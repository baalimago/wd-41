package serve

import "testing"

func Test_Setup(t *testing.T) {
	t.Run("it should set relPath to second argument", func(t *testing.T) {
		want := "post"
		c := command{
			relPath: "pre",
		}
		given := []string{want}
		err := c.Flagset().Parse(given)
		if err != nil {
			t.Fatalf("failed to parse flagset: %v", err)
		}
		c.Setup()
		got := c.relPath
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
	t.Run("it should set port arg", func(t *testing.T) {
		want := 9090
		c := command{}
		givenArgs := []string{"-port", "9090"}
		err := c.Flagset().Parse(givenArgs)
		if err != nil {
			t.Fatalf("failed to parse flagset: %v", err)
		}
		err = c.Setup()
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}

		got := *c.port
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}
