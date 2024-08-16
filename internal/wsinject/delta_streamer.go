package wsinject

import (
	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"golang.org/x/net/websocket"
)

// WsHandler echoes received messages back to the client.
func (fs *Fileserver) WsHandler(ws *websocket.Conn) {
	reloadChan := make(chan string)
	dieChan := make(chan struct{})
	go func() {
		ancli.PrintfOK("new websocket connection: '%v'", ws.Config().Origin)
		for {
			select {
			case pageToReload := <-reloadChan:
				err := websocket.Message.Send(ws, pageToReload)
				if err != nil {
					// Exit on error
					ancli.PrintfErr("ws: failed to send message via ws: %v", err)
					dieChan <- struct{}{}
				}
			}
		}
	}()

	ancli.PrintOK("Listening to file changes on pageReloadChan")
	for {
		select {
		case pageToReload := <-fs.pageReloadChan:
			reloadChan <- pageToReload
		case <-dieChan:
			ancli.PrintOK("websocket handler goes byebye")
			return
		}
	}
}
