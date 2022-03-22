package commands

import (
	"github.com/jimeh/evm/manager"
	"github.com/spf13/cobra"
)

func NewConfig(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:       "config",
		Short:     "Show evm environment/setup details",
		Aliases:   []string{"env", "info"},
		ValidArgs: []string{},
		RunE:      configRunE(mgr),
	}

	cmd.Flags().StringP("format", "f", "", "output format (yaml or json)")

	return cmd, nil
}

func configRunE(mgr *manager.Manager) runEFunc {
	return func(cmd *cobra.Command, _ []string) error {
		format := flagString(cmd, "format")

		return render(cmd.OutOrStdout(), format, mgr.Config)
	}
}
