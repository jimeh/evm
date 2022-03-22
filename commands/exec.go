package commands

import (
	"bytes"
	"errors"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/jimeh/evm/manager"
	"github.com/spf13/cobra"
)

func NewExec(mgr *manager.Manager) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "exec <binary> [args]...",
		Short: "Execute named binary from current " +
			"Emacs version",
		Args:                  cobra.MinimumNArgs(1),
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		ValidArgsFunction:     execValidArgs(mgr),
		RunE:                  execRunE(mgr),
	}

	return cmd, nil
}

func execRunE(mgr *manager.Manager) runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		program := args[0]
		args = args[1:]

		err := mgr.Exec(ctx, program, args)
		if err != nil {
			if errors.Is(err, manager.ErrBinNotFound) {
				var versions []*manager.Version
				versions, err = mgr.FindBin(ctx, program)
				if err != nil {
					return err
				}

				return newExecOtherVersionsError(&execOtherVersionsData{
					Name:        program,
					Current:     mgr.CurrentVersion(),
					AvailableIn: versions,
				})
			}

			if errors.Is(err, manager.ErrNoCurrentVersion) {
				return newExecNoCurrentVersionError()
			}

			return err
		}

		return nil
	}
}

func execValidArgs(mgr *manager.Manager) validArgsFunc {
	return func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		var r []string

		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		version, err := mgr.Get(cmd.Context(), mgr.CurrentVersion())
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		for _, bin := range version.Binaries {
			base := filepath.Base(bin)
			if toComplete == "" || strings.HasPrefix(base, toComplete) {
				r = append(r, base)
			}
		}

		return r, cobra.ShellCompDirectiveNoFileComp
	}
}

func newExecNoCurrentVersionError() error {
	return errors.New(`No current Emacs version is set.

List all installed versions with: evm list
Change version with: evm use <version>`,
	)
}

type execOtherVersionsData struct {
	Name        string
	Current     string
	AvailableIn []*manager.Version
}

func newExecOtherVersionsError(data *execOtherVersionsData) error {
	var buf bytes.Buffer
	err := execOtherVersionsTemplate.Execute(&buf, data)
	if err != nil {
		return err
	}

	return errors.New(buf.String())
}

var execOtherVersionsTemplate = template.Must(template.New("other").Parse(
	`{{ if gt (len .AvailableIn) 0 -}}
Executable "{{.Name}}" is not available in the current Emacs version ({{.Current}}).

"{{.Name}}" is available in the following Emacs versions:
{{- range .AvailableIn }}
  - {{ .Version }}
{{- end -}}
{{ else -}}
Executable "{{.Name}}" is not available in any installed Emacs version.
{{- end }}

Change version with: evm use <version>
List all installed versions with: evm list`,
))
