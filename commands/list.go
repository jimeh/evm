package commands

import (
	"strings"

	"github.com/jimeh/evm/manager"
	"github.com/jimeh/go-render"
	"github.com/spf13/cobra"
)

func NewList(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "list",
		Short: "List all Emacs versions found in " +
			"`$EVM_ROOT/versions/*'",
		Aliases:           []string{"ls", "versions"},
		Args:              cobra.ExactArgs(0),
		ValidArgsFunction: noValidArgs,
		RunE:              listRunE(mgr),
	}

	cmd.Flags().StringP(
		"format", "f", "text", "output format ,\"text\", \"yaml\", or \"json\"",
	)

	return cmd, nil
}

func listRunE(mgr *manager.Manager) runEFunc {
	return func(cmd *cobra.Command, _ []string) error {
		format := flagString(cmd, "format")

		versions, err := mgr.List(cmd.Context())
		if err != nil {
			return err
		}

		output := &listOutput{
			Current: listOutputCurrent{
				Version: mgr.CurrentVersion(),
				SetBy:   mgr.CurrentSetBy(),
			},
			Versions: versions,
		}

		return render.Pretty(cmd.OutOrStdout(), format, output)
	}
}

type listOutput struct {
	Current  listOutputCurrent  `yaml:"current" json:"current"`
	Versions []*manager.Version `yaml:"versions" json:"versions"`
}

type listOutputCurrent struct {
	Version string `yaml:"version" json:"version"`
	SetBy   string `yaml:"set_by,omitempty" json:"set_by,omitempty"`
}

func (lo *listOutput) String() string {
	buf := &strings.Builder{}

	for _, ver := range lo.Versions {
		if lo.Current.Version == ver.Version {
			buf.WriteString("* ")
		} else {
			buf.WriteString("  ")
		}

		buf.WriteString(ver.Version)
		if lo.Current.Version == ver.Version && lo.Current.SetBy != "" {
			buf.WriteString(" (set by " + lo.Current.SetBy + ")")
		}

		buf.WriteByte('\n')
	}

	return buf.String()
}
