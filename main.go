package main

import (
	"context"
	"os"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"github.com/baalimago/go_away_boilerplate/pkg/shutdown"
	"github.com/baalimago/wd-41/cmd"
	"github.com/baalimago/wd-41/cmd/serve"
	"github.com/baalimago/wd-41/cmd/version"
)

var commands = map[string]cmd.Command{
	"s|serve":   serve.Command(),
	"v|version": version.Command(),
}

const usage = `== Web Development 41 == 

This tool is designed to enable live reload for statically hosted web development.
It injects a websocket script in a mirrored version of html pages
and uses the fsnotify (cross-platform 'inotify' wrapper) package to detect filechanges.
On filechanges, the websocket will trigger a reload of the page.

The 41 (formerly "40", before I got spooked by potential lawyers) is only 
to enable rust-repellant properties.

Commands:
%v`

func main() {
	ancli.Newline = true
	ancli.SetupSlog()
	version.Name = "wd-41"
	ctx, cancel := context.WithCancel(context.Background())
	exitCodeChan := make(chan int, 1)
	go func() {
		exitCodeChan <- cmd.Run(ctx, os.Args, commands, usage)
		cancel()
	}()
	shutdown.MonitorV2(ctx, cancel)
	os.Exit(<-exitCodeChan)
}
