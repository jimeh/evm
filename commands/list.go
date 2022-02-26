package commands

import (
	"io"

	"github.com/jimeh/evm/manager"
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

	cmd.Flags().StringP("format", "f", "", "output format (yaml or json)")

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
			Current:  mgr.CurrentVersion(),
			SetBy:    mgr.CurrentSetBy(),
			Versions: versions,
		}

		return render(cmd.OutOrStdout(), format, output)
	}
}

type listOutput struct {
	Current  string             `yaml:"current" json:"current"`
	SetBy    string             `yaml:"current_set_by,omitempty" json:"current_set_by,omitempty"`
	Versions []*manager.Version `yaml:"versions" json:"versions"`
}

func (lr *listOutput) WriteTo(w io.Writer) (int64, error) {
	var b []byte

	for _, ver := range lr.Versions {
		if lr.Current == ver.Version {
			b = append(b, []byte("* ")...)
		} else {
			b = append(b, []byte("  ")...)
		}

		b = append(b, []byte(ver.Version)...)
		if lr.Current == ver.Version && lr.SetBy != "" {
			b = append(b, []byte(" (set by "+lr.SetBy+")")...)
		}

		b = append(b, byte('\n'))
	}

	n, err := w.Write(b)

	return int64(n), err
}
