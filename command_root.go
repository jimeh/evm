package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func rootCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "A simple and opinionated Emacs Version Manager and build tool.",
	}

	mode := os.Getenv("EVM_MODE")
	if mode != "system" {
		mode = "user"
	}

	viper.SetDefault("mode", mode)

	rootDir := filepath.Join(string(os.PathSeparator), "opt", "evm")
	if mode == "user" {
		rootDir = filepath.Join("$HOME", ".evm")
	}

	cmd.PersistentFlags().String("root", rootDir, "Root directory")
	viper.SetDefault("path.root", rootDir)
	err := viper.BindPFlag(
		"path.root", cmd.PersistentFlags().Lookup("root"),
	)
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("path.root", "EVM_ROOT")
	if err != nil {
		return nil, err
	}

	viper.SetDefault("path.shims", filepath.Join("$EVM_ROOT", "shims"))
	viper.SetDefault("path.sources", filepath.Join("$EVM_ROOT", "sources"))
	viper.SetDefault("path.versions", filepath.Join("$EVM_ROOT", "versions"))

	infoCmd, err := configCommand()
	if err != nil {
		return nil, err
	}

	listCmd, err := listCommand()
	if err != nil {
		return nil, err
	}

	useCmd, err := useCommand()
	if err != nil {
		return nil, err
	}

	execCmd, err := execCommand()
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(
		infoCmd,
		listCmd,
		useCmd,
		execCmd,
	)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("evm")
	viper.AutomaticEnv()

	return cmd, nil
}
