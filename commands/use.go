package commands

import (
	"strings"

	"github.com/jimeh/evm/manager"
	"github.com/spf13/cobra"
)

func NewUse(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:               "use <version>",
		Short:             "Switch to a specific version",
		Aliases:           []string{"activate", "switch"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: useValidArgs(mgr),
		RunE:              useRunE(mgr),
	}

	return cmd, nil
}

func useRunE(mgr *manager.Manager) runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		return mgr.Use(cmd.Context(), args[0])
	}
}

func useValidArgs(mgr *manager.Manager) validArgsFunc {
	return func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		versions, err := mgr.List(cmd.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var r []string
		for _, ver := range versions {
			if toComplete == "" || strings.HasPrefix(ver.Version, toComplete) {
				r = append(r, ver.Version)
			}
		}

		return r, cobra.ShellCompDirectiveNoFileComp
	}
}
