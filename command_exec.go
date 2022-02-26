package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"syscall"
	"text/template"

	"github.com/spf13/cobra"
)

func execCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "exec <binary> [args]...",
		Short: "Execute named binary from current " +
			"Emacs version",
		Args:                  cobra.MinimumNArgs(1),
		SilenceUsage:          true,
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		ValidArgsFunction:     execValidArgs,
		RunE:                  execRunE,
	}

	return cmd, nil
}

func execRunE(cmd *cobra.Command, args []string) error {
	conf, err := getConfig()
	if err != nil {
		return err
	}

	version, err := newEmacsVersion(conf, conf.Current.Version)
	if err != nil {
		return err
	}

	if bin, ok := version.BinPath(args[0]); ok {
		select {
		case <-cmd.Context().Done():
			return cmd.Context().Err()
		default:
		}

		execArgs := append([]string{bin}, args[1:]...)
		execEnv := os.Environ()
		for i := 0; i < len(execEnv); i++ {
			if strings.HasPrefix(execEnv[i], "PATH=") {
				execEnv[i] = "PATH=" + version.Bin + ":" + execEnv[i][5:]
			}
		}

		return syscall.Exec(bin, execArgs, execEnv)
	}

	versions, err := newEmacsVersions(conf)
	if err != nil {
		return err
	}

	var availableIn []string
	for _, ev := range versions {
		if _, ok := ev.BinPath(args[0]); ok {
			availableIn = append(availableIn, ev.Version)
		}
	}

	var buf bytes.Buffer
	err = execOtherVersionsTemplate.Execute(&buf, &execOtherVersionsData{
		Name:        args[0],
		Current:     conf.Current.Version,
		AvailableIn: availableIn,
	})
	if err != nil {
		return err
	}

	return errors.New(buf.String())
}

func execValidArgs(
	_ *cobra.Command,
	args []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	var r []string

	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	conf, err := getConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	version, err := newEmacsVersion(conf, conf.Current.Version)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	for _, bin := range version.Binaries {
		if toComplete == "" || strings.HasPrefix(bin, toComplete) {
			r = append(r, bin)
		}
	}

	return r, cobra.ShellCompDirectiveNoFileComp
}

type execOtherVersionsData struct {
	Name        string
	Current     string
	AvailableIn []string
}

var execOtherVersionsTemplate = template.Must(template.New("other").Parse(
	`{{ if gt (len .AvailableIn) 0 -}}
Current Emacs version ({{.Current}}) does not have a "{{.Name}}" executable.

"{{.Name}}" is available in the following Emacs versions:
{{- range .AvailableIn }}
  - {{ . }}
{{- end -}}
{{ else -}}
"{{.Name}}" executable is not available in any installed Emacs version.
{{- end }}

Change version with: evm use <version>
List all installed versions with: evm list`,
))
