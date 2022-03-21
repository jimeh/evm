package commands

import (
	"github.com/jimeh/evm/manager"
	"github.com/spf13/cobra"
)

func NewEvm(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "A simple and opinionated Emacs Version Manager and build tool",
	}

	cmd.PersistentFlags().StringP(
		"log-level", "l", "info",
		"one of: trace, debug, info, warn, error, fatal, panic",
	)

	configCmd, err := NewConfig(mgr)
	if err != nil {
		return nil, err
	}

	listCmd, err := NewList(mgr)
	if err != nil {
		return nil, err
	}

	useCmd, err := NewUse(mgr)
	if err != nil {
		return nil, err
	}

	rehashCmd, err := NewRehash(mgr)
	if err != nil {
		return nil, err
	}

	execCmd, err := NewExec(mgr)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(
		configCmd,
		listCmd,
		useCmd,
		rehashCmd,
		execCmd,
	)

	return cmd, nil
}
