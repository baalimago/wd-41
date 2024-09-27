package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/baalimago/wd-41/cmd/serve"
	"github.com/baalimago/wd-41/cmd/version"
)

var commands = map[string]Command{
	"s|serve":   serve.Command(),
	"v|version": version.Command(),
}

func Parse(args []string) (Command, error) {
	if len(args) == 1 {
		return nil, ErrNoArgs
	}
	cmdCandidate := ""
	for _, arg := range args[1:] {
		if isHelp(arg) {
			return nil, ErrHelpful
		}
		isFlag := strings.HasPrefix(arg, "-")
		if isFlag {
			continue
		}
		// Break on first non-flag
		cmdCandidate = arg
		break
	}
	for cmdNameWithShortcut, cmd := range commands {
		for _, cmdName := range strings.Split(cmdNameWithShortcut, "|") {
			exists := cmdName == cmdCandidate
			if exists {
				return cmd, nil
			}
		}
	}

	return nil, ArgNotFoundError(cmdCandidate)
}

func isHelp(s string) bool {
	return s == "-h" || s == "-help" || s == "h" || s == "help"
}

func formatCommandDescriptions() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for name, cmd := range commands {
		fmt.Fprintf(w, "\t%v\t%v\n", name, cmd.Describe())
	}
	w.Flush()
	return buf.String()
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

func PrintUsage() {
	fmt.Printf(usage, formatCommandDescriptions())
}
