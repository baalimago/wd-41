package serve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
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
	ancli.Newline = true
	cmd := command{}
	cmd.fileserver = &mockFileServer{}
	fs := cmd.Flagset()
	fs.Parse([]string{"--port=8081", "--wsPort=/test-ws"})

	err := cmd.Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the server in a separate goroutine
	go func() {
		err := cmd.Run(ctx)
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// Test if the HTTP server is working
	resp, err := http.Get("http://localhost:8081/")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got: %v", resp.Status)
	}

	// Test the websocket handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := websocket.Server{Handler: cmd.fileserver.WsHandler}
		s.ServeHTTP(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	ws, err := websocket.Dial(wsURL+"/test-ws", "", "http://localhost/")
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer ws.Close()

	// Cleanup
	cancel()
	time.Sleep(time.Second)
}
