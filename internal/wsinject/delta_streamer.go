package wsinject

import (
	"fmt"
	"math/rand"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"github.com/baalimago/go_away_boilerplate/pkg/threadsafe"
	"golang.org/x/net/websocket"
)

// WsHandler echoes received messages back to the client.
func (fs *Fileserver) WsHandler(ws *websocket.Conn) {
	reloadChan := make(chan string)
	killChan := make(chan struct{})
	name := "ws-" + fmt.Sprintf("%v", rand.Int())

	go func() {
		ancli.PrintfOK("new websocket connection: '%v'", ws.Config().Origin)
		for {
			select {
			case pageToReload, ok := <-reloadChan:
				if !ok {
					killChan <- struct{}{}
				}
				err := websocket.Message.Send(ws, pageToReload)
				if err != nil {
					// Exit on error
					ancli.PrintfErr("ws: failed to send message via ws: %v", err)
					killChan <- struct{}{}
				}
			}
		}
	}()

	ancli.PrintOK("Listening to file changes on pageReloadChan")
	fs.registerWs(name, reloadChan)
	for {
		select {
		case <-killChan:
			ancli.PrintOK("websocket disconnected")
			fs.deregisterWs(name)
			err := ws.WriteClose(1005)
			if err != nil {
				ancli.PrintfErr("ws-listener: '%v' got err when writeclosing: %v", name, err)
			}
			err = ws.Close()
			if err != nil {
				ancli.PrintfErr("ws-listener: '%v' got err when closing: %v", name, err)
			}

			return
		}
	}
}

func (fs *Fileserver) registerWs(name string, c chan string) {
	if !threadsafe.Read(fs.wsDispatcherStartedMu, fs.wsDispatcherStarted) {
		go fs.wsDispatcherStart()
		threadsafe.Write(fs.wsDispatcherStartedMu, true, fs.wsDispatcherStarted)
	}
	ancli.PrintfNotice("registering: '%v'", name)
	fs.wsDispatcher.Store(name, c)
}

func (fs *Fileserver) deregisterWs(name string) {
	fs.wsDispatcher.Delete(name)
}

func (fs *Fileserver) wsDispatcherStart() {
	for {
		select {
		case pageToReload, ok := <-fs.pageReloadChan:
			if !ok {
				ancli.PrintNotice("stopping wsDispatcher")
				fs.wsDispatcher.Range(func(key, value any) bool {
					ancli.PrintfNotice("sending to: '%v'", key)
					wsWriterChan := value.(chan string)
					// Close chan to stop the wsRoutine
					close(wsWriterChan)
					return true
				})
				return
			}
			ancli.PrintfNotice("got update: '%v'", pageToReload)
			fs.wsDispatcher.Range(func(key, value any) bool {
				ancli.PrintfNotice("sending to: '%v'", key)
				wsWriterChan := value.(chan string)
				wsWriterChan <- pageToReload
				return true
			})
		}
	}
}
