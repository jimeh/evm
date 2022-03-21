package commands

import (
	"strings"

	"github.com/jimeh/evm/manager"
	"github.com/spf13/cobra"
)

func NewRehash(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:               "rehash [<version>...]",
		Short:             "Update shims for all or specific versions",
		Aliases:           []string{"reshim"},
		ValidArgsFunction: rehashValidArgs(mgr),
		RunE:              WithPrettyLogging(rehashRunE(mgr)),
	}

	return cmd, nil
}

func rehashRunE(mgr *manager.Manager) runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if len(args) > 0 {
			return mgr.RehashVersions(ctx, args)
		}

		return mgr.RehashAll(ctx)
	}
}

func rehashValidArgs(mgr *manager.Manager) validArgsFunc {
	return func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		versions, err := mgr.List(cmd.Context())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var r []string
		for _, ver := range versions {
			if stringsContains(args, ver.Version) {
				continue
			}

			if toComplete == "" || strings.HasPrefix(ver.Version, toComplete) {
				r = append(r, ver.Version)
			}
		}

		return r, cobra.ShellCompDirectiveNoFileComp
	}
}
