package serve

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

type command struct {
	relPath string
	binPath string
	port    *int
	flagset *flag.FlagSet
}

func Command() *command {
	r, _ := os.Executable()
	return &command{
		binPath: r,
	}
}

func (c *command) Setup() error {
	relPath := ""
	if len(c.flagset.Args()) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get exec path: %w", err)
		}
		relPath = wd
	} else {
		relPath = c.flagset.Arg(0)
	}
	c.relPath = path.Clean(relPath)
	return nil
}

func (c *command) Run(ctx context.Context) error {
	h := http.FileServer(http.Dir(c.relPath))
	s := http.Server{
		Addr:    fmt.Sprintf(":%v", *c.port),
		Handler: h,
	}
	errChan := make(chan error, 1)
	go func() {
		ancli.PrintfOK("now serving directory: '%v' on port: '%v'", c.relPath, *c.port)
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()
	select {
	case <-ctx.Done():
	case serveErr := <-errChan:
		return serveErr
	}
	ancli.PrintNotice("initiating webserver graceful shutdown")
	s.Shutdown(ctx)
	ancli.PrintOK("shutdown complete")
	return nil
}

func (c *command) Help() string {
	return "Serve some filesystem. Set the directory as the second argument: wd-40 serve <dir>. If omitted, current wd will be used."
}

func (c *command) Describe() string {
	return fmt.Sprintf("a webserver. Usage: '%v serve <path>'. If <path> is left unfilled, current pwd will be used.", c.binPath)
}

func (c *command) Flagset() *flag.FlagSet {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	c.port = fs.Int("port", 8080, "port to serve http server on")
	c.flagset = fs
	return fs
}
