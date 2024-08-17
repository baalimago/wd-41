package wsinject

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"golang.org/x/net/websocket"
)

func TestWsHandler(t *testing.T) {

	ancli.Newline = true
	setup := func(t *testing.T) (*Fileserver, *websocket.Config, *httptest.Server) {
		t.Helper()
		started := false
		fs := &Fileserver{
			pageReloadChan:        make(chan string),
			wsDispatcher:          sync.Map{},
			wsDispatcherStarted:   &started,
			wsDispatcherStartedMu: &sync.Mutex{},
		}

		server := httptest.NewServer(websocket.Handler(fs.WsHandler))

		port := strings.Replace(server.URL, "http://127.0.0.1:", "", -1)
		wsConfig, err := websocket.NewConfig(fmt.Sprintf("ws://localhost:%v", port), "ws://localhost/")
		if err != nil {
			t.Fatalf("Failed to create WebSocket config: %v", err)
		}

		return fs, wsConfig, server
	}

	t.Run("it should send messages posted on pageReloadChan", func(t *testing.T) {
		fs, wsConfig, testServer := setup(t)

		ws, err := websocket.DialConfig(wsConfig)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		t.Cleanup(func() {
			testServer.Close()
			ws.Close()
		})

		go func() {
			fs.pageReloadChan <- "test message"
		}()

		var msg string
		err = websocket.Message.Receive(ws, &msg)
		if err != nil {
			t.Fatalf("Failed to receive message: %v", err)
		}

		if msg != "test message" {
			t.Fatalf("Expected 'test message', got: %v", msg)
		}

		close(fs.pageReloadChan)
		select {
		case <-time.After(time.Second):
			t.Fatal("Expected the WebSocket to be closed")
		case <-fs.pageReloadChan:
		}
	})

	t.Run("it should handle multiple connections at once", func(t *testing.T) {
		fs, wsConfig, testServer := setup(t)

		mockWebClient0, err := websocket.DialConfig(wsConfig)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}

		mockWebClient1, err := websocket.DialConfig(wsConfig)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}

		t.Cleanup(func() {
			mockWebClient0.Close()
			mockWebClient1.Close()
			testServer.Close()
		})

		mu := &sync.Mutex{}
		go func() {
			mu.Lock()
			defer mu.Unlock()
			fs.pageReloadChan <- "test message"
		}()

		gotMsgChan := make(chan string)
		for _, wsClient := range []*websocket.Conn{mockWebClient0, mockWebClient1} {
			go func(wsClient *websocket.Conn) {
				for {
					var msg string
					websocket.Message.Receive(wsClient, &msg)
					gotMsgChan <- msg
				}
			}(wsClient)
		}
		want := 0
		for want != 2 {
			select {
			case <-time.After(time.Second):
				t.Fatal("failed to recieve data from websocket")
			case got := <-gotMsgChan:
				want += 1
				t.Logf("got message from mocked ws client: %v", got)
			}
		}

		mu.Lock()
		close(fs.pageReloadChan)
		mu.Unlock()
	})

}
