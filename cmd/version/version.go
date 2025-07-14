package version

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"runtime/debug"
)

// Set with buildflag if built in pipeline and not using go install
var (
	BuildVersion  = ""
	BuildChecksum = ""
)

type command struct {
	getVersionCmd func() (*debug.BuildInfo, bool)
}

// Describe the version *command
func (c *command) Describe() string {
	return "print the version of wd-41"
}

// Flagset for version, currently empty
func (c *command) Flagset() *flag.FlagSet {
	return flag.NewFlagSet("version", flag.ExitOnError)
}

// Help by printing out help
func (c *command) Help() string {
	return "Print the version of wd-41"
}

// Run the *command, printing the version using either the debugbuild or tagged version
func (c *command) Run(context.Context) error {
	bi, ok := c.getVersionCmd()
	if !ok {
		return errors.New("failed to read build info")
	}
	version := bi.Main.Version
	checksum := bi.Main.Sum
	if version == "" || version == "(devel)" {
		version = BuildVersion
	}
	if checksum == "" {
		checksum = BuildChecksum
	}
	fmt.Printf("version: %v, go version: %v, checksum: %v\n", version, bi.GoVersion, checksum)
	return nil
}

// Setup the *command
func (c *command) Setup() error {
	c.getVersionCmd = debug.ReadBuildInfo
	return nil
}

func Command() *command {
	return &command{}
}
