package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
)

type Command interface {
	Setup(context.Context) error

	// Run and block until context cancel
	Run(context.Context) error

	// Help by printing a usage string. Currently not used anywhere.
	Help() string

	// Describe the command shortly
	Describe() string

	// Flagset which defines the flags for the command
	Flagset() *flag.FlagSet
}

type (
	ArgParser    func([]string, map[string]Command) (Command, func(string, map[string]Command) string, error)
	UsagePrinter func()
)

type ArgNotFoundError string

func (e ArgNotFoundError) Error() string {
	return fmt.Sprintf("'%v' is not a valid argument\n", string(e))
}

var ErrNoArgs = errors.New("no arguments found")
