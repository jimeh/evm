package main

import (
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func listCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "list",
		Short: "List all Emacs versions found in " +
			"`$EVM_ROOT/versions/*'",
		Aliases:           []string{"ls", "versions"},
		Args:              cobra.ExactArgs(0),
		ValidArgsFunction: listValidArgs,
		RunE:              listRunE,
	}

	cmd.Flags().StringP("format", "f", "", "output format (yaml or json)")
	err := viper.BindPFlag("list.format", cmd.Flags().Lookup("format"))
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func listRunE(cmd *cobra.Command, _ []string) error {
	conf, err := getConfig()
	if err != nil {
		return err
	}

	format := viper.GetString("list.format")

	versions, err := newEmacsVersions(conf)
	if err != nil {
		return err
	}

	results := &listResults{
		Current:  conf.Current.Version,
		Versions: versions,
	}

	return render(cmd.OutOrStdout(), format, results)
}

func listValidArgs(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

type listResults struct {
	Current  string        `yaml:"current" json:"current"`
	Versions emacsVersions `yaml:"versions" json:"versions"`
}

func (lr *listResults) WriteTo(w io.Writer) (int64, error) {
	var b []byte

	for _, ver := range lr.Versions {
		if lr.Current == ver.Version {
			b = append(b, []byte("* ")...)
		} else {
			b = append(b, []byte("  ")...)
		}

		b = append(b, []byte(ver.Version)...)
		b = append(b, byte('\n'))
	}

	n, err := w.Write(b)

	return int64(n), err
}
