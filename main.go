package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"github.com/baalimago/go_away_boilerplate/pkg/shutdown"
	"github.com/baalimago/wd-40/cmd"
)

func printHelp(command cmd.Command, err error, printUsage cmd.UsagePrinter) int {
	var notValidArg cmd.ArgNotFoundError
	if errors.As(err, &notValidArg) {
		ancli.PrintErr(err.Error())
		printUsage()
	} else if errors.Is(err, cmd.ErrNoArgs) {
		printUsage()
	} else if errors.Is(err, cmd.ErrHelpful) {
		if command != nil {
			fmt.Println(command.Help())
		} else {
			printUsage()
		}
		return 0
	} else {
		ancli.PrintfErr("unknown error: %v", err.Error())
	}
	return 1
}

func run(ctx context.Context, args []string, parseArgs cmd.ArgParser) int {
	command, err := parseArgs(args)
	if err != nil {
		return printHelp(command, err, cmd.PrintUsage)
	}
	fs := command.Flagset()
	var cmdArgs []string
	if len(args) > 2 {
		cmdArgs = args[2:]
	}
	err = fs.Parse(cmdArgs)
	if err != nil {
		ancli.PrintfErr("failed to parse flagset: %v", err.Error())
		return 1
	}

	err = command.Setup()
	if err != nil {
		ancli.PrintfErr("failed to setup command: %v", err.Error())
		return 1
	}

	err = command.Run(ctx)
	if err != nil {
		ancli.PrintfErr("failed to run %v", err.Error())
		return 1
	}
	return 0
}

func main() {
	ancli.Newline = true
	ancli.SlogIt = true
	ancli.SetupSlog()
	ctx, cancel := context.WithCancel(context.Background())
	exitCodeChan := make(chan int, 1)
	go func() {
		exitCodeChan <- run(ctx, os.Args, cmd.Parse)
		cancel()
	}()
	shutdown.MonitorV2(ctx, cancel)
	os.Exit(<-exitCodeChan)
}
