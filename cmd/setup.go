package cmd

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func Run(ctx context.Context, args []string, commands map[string]Command, usage string) int {
	command, err := parse(args, commands)
	if err != nil {
		return printHelp(command, err, func() {
			ancli.Okf("%v", getUsage(usage, commands))
		})
	}

	err = command.Setup(context.Background())
	if err != nil {
		ancli.Errf("failed to setup command: %v", err.Error())
		return 1
	}

	err = command.Run(ctx)
	if err != nil {
		ancli.Errf("failed to run: %v", err.Error())
		return 1
	}
	return 0
}

func parse(args []string, commands map[string]Command) (Command, error) {
	if len(args) == 1 {
		return nil, ErrNoArgs
	}
	// Strip binary from args to find first argument
	args = args[1:]
	var command Command
	cmdCandidate := ""
	cmdIdx := -1
	for idx, arg := range args {
		isFlag := strings.HasPrefix(arg, "-")
		if isFlag {
			continue
		}
		// Break on first non-flag
		cmdCandidate = arg
		cmdIdx = idx
		break
	}

OUTER:
	for cmdNameWithShortcut, cmd := range commands {
		for _, cmdName := range strings.Split(cmdNameWithShortcut, "|") {
			if cmdName == cmdCandidate {
				command = cmd
				break OUTER
			}
		}
	}

	if command == nil {
		return nil, ArgNotFoundError(cmdCandidate)
	}

	args = append(args[:cmdIdx], args[cmdIdx+1:]...)

	fs := command.Flagset()
	// Strip found command
	err := fs.Parse(args)
	if err != nil {
		return command, fmt.Errorf("failed to parse flagset: %w", err)
	}

	return command, nil
}

func printHelp(command Command, err error, printUsage UsagePrinter) int {
	var notValidArg ArgNotFoundError
	if errors.As(err, &notValidArg) {
		ancli.Errf("%v", err.Error())
	} else if errors.Is(err, ErrNoArgs) {
	} else if errors.Is(err, flag.ErrHelp) && command != nil {
		ancli.Noticef("[command help]: %v", command.Help())
		return 0
	} else {
		ancli.Errf("unknown error: %v", err.Error())
	}
	printUsage()
	return 1
}

func formatCommandDescriptions(commands map[string]Command) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for name, cmd := range commands {
		fmt.Fprintf(w, "\t%v\t%v\n", name, cmd.Describe())
	}
	w.Flush()
	return buf.String()
}

func getUsage(usage string, cmds map[string]Command) string {
	return fmt.Sprintf(usage, formatCommandDescriptions(cmds))
}
