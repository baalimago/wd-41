package serve

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/testboil"
	"golang.org/x/net/websocket"
)

func Test_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	t.Run("it should set masterPath to second argument", func(t *testing.T) {
		want := tmpDir
		c := command{
			masterPath: "pre",
		}
		given := []string{want}
		err := c.Flagset().Parse(given)
		if err != nil {
			t.Fatalf("failed to parse flagset: %v", err)
		}
		c.Setup()
		got := c.masterPath
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

	t.Run("it should set cacheControl arg", func(t *testing.T) {
		want := "test"
		c := command{}
		givenArgs := []string{"-cacheControl", want}
		err := c.Flagset().Parse(givenArgs)
		if err != nil {
			t.Fatalf("failed to parse flagset: %v", err)
		}
		err = c.Setup()
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}

		got := *c.cacheControl
		if got != want {
			t.Fatalf("expected: %v, got: %v", want, got)
		}
	})
}

type mockFileServer struct{}

func (m *mockFileServer) Setup(pathToMaster string) (string, error) {
	return "/mock/mirror/path", nil
}

func (m *mockFileServer) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (m *mockFileServer) WsHandler(ws *websocket.Conn) {}

func TestRun(t *testing.T) {
	setup := func() command {
		cmd := command{}
		cmd.fileserver = &mockFileServer{}
		fs := cmd.Flagset()
		fs.Parse([]string{"--port=8081", "--wsPort=/test-ws"})

		err := cmd.Setup()
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
		return cmd
	}

	t.Run("it should setup websocket handler on wsPort", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())

		ready := make(chan struct{})
		go func() {
			close(ready)
			err := cmd.Run(ctx)
			if err != nil {
				t.Errorf("Run returned error: %v", err)
			}
		}()

		t.Cleanup(ctxCancel)

		<-ready
		// Test if the HTTP server is working
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			t.Fatalf("Failed to send GET request: %v", err)
		}
		t.Cleanup(func() { resp.Body.Close() })

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK, got: %v", resp.Status)
		}

		// Test the websocket handler
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := websocket.Server{Handler: cmd.fileserver.WsHandler}
			s.ServeHTTP(w, r)
		}))

		t.Cleanup(func() { server.Close() })

		wsURL := "ws" + server.URL[len("http"):]
		ws, err := websocket.Dial(wsURL+"/test-ws", "", "http://localhost/")
		if err != nil {
			t.Fatalf("websocket dial failed: %v", err)
		}
		t.Cleanup(func() { ws.Close() })
	})

	t.Run("it should respond with correct cache control", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)
		want := "test"
		port := 13337
		cmd.cacheControl = &want
		cmd.port = &port

		ready := make(chan struct{})
		go func() {
			close(ready)
			err := cmd.Run(ctx)
			if err != nil {
				t.Errorf("Run returned error: %v", err)
			}
		}()
		<-ready
		time.Sleep(time.Millisecond)
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v", port))
		if err != nil {
			t.Fatal(err)
		}
		got := resp.Header.Get("Cache-Control")
		testboil.FailTestIfDiff(t, got, want)
	})

	t.Run("it should serve with tls if cert is specified", func(t *testing.T) {
		cmd := setup()
		ctx, ctxCancel := context.WithCancel(context.Background())
		t.Cleanup(ctxCancel)
		port := 13337
		cmd.port = &port
		cmd.certificatePath = "TODO"

		ready := make(chan struct{})
		go func() {
			close(ready)
			err := cmd.Run(ctx)
			if err != nil {
				t.Errorf("Run returned error: %v", err)
			}
		}()
		<-ready
		time.Sleep(time.Millisecond)
		resp, err := http.Get(fmt.Sprintf("http://localhost:%v", port))
		if err != nil {
			t.Fatal(err)
		}
		got := resp.Header.Get("Cache-Control")
		testboil.FailTestIfDiff(t, got, want)
	})
}
