package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:       "config",
		Short:     "Show evm environment/setup details.",
		Aliases:   []string{"env", "info"},
		ValidArgs: []string{},
		RunE:      configRunE,
	}

	cmd.Flags().StringP("format", "f", "", "output format (yaml or json)")
	err := viper.BindPFlag("info.format", cmd.Flags().Lookup("format"))
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func configRunE(cmd *cobra.Command, _ []string) error {
	conf, err := getConfig()
	if err != nil {
		return err
	}

	format := viper.GetString("info.format")

	return render(cmd.OutOrStdout(), format, conf)
}
