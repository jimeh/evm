package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jimeh/evm/commands"
	"github.com/jimeh/evm/manager"
)

func main() {
	mgr, err := manager.New(nil)
	if err != nil {
		fatal(err)
	}

	cmd, err := commands.NewEvm(mgr)
	if err != nil {
		fatal(err)
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	err = cmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	os.Exit(1)
}
