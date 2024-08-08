package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/baalimago/wd-40/cmd/serve"
)

var commands = map[string]Command{
	"serve": serve.Command(),
}

func Parse(args []string) (Command, error) {
	if len(args) == 1 {
		return nil, NoArgsError
	}
	cmdCandidate := ""
	for _, arg := range args[1:] {
		isFlag := strings.HasPrefix(arg, "-")
		if isFlag {
			continue
		}
		// Break on first non-flag
		cmdCandidate = arg
		break
	}
	cmd, exists := commands[cmdCandidate]
	if !exists {
		return nil, ArgNotFoundError(cmdCandidate)
	}
	return cmd, nil
}

func formatCommandDescriptions() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for name, cmd := range commands {
		fmt.Fprintf(w, "\t%v -\t%v\n", name, cmd.Describe())
	}
	w.Flush()
	return buf.String()
}

const usage = `== Web Development 40 == 

This tool is designed to ease the web devleopment hassles by
automatically attaching a script which sets up websocket. Through
this websocket, dynamic hot reloads of file changes are streamed.. somehow
haven't figured out that part yet.

The 40 is only to enable rust-repellant properties.

Commands:
%v`

func PrintUsage() {
	fmt.Printf(usage, formatCommandDescriptions())
}
