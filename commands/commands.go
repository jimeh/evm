package commands

import "github.com/spf13/cobra"

type runEFunc func(cmd *cobra.Command, _ []string) error

type validArgsFunc func(
	cmd *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective)

func noValidArgs(
	_ *cobra.Command,
	_ []string,
	_ string,
) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func flagString(cmd *cobra.Command, name string) string {
	var r string

	if f := cmd.Flag(name); f != nil {
		r = f.Value.String()
	}

	return r
}

func stringsContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}
