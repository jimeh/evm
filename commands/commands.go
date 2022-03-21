package commands

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type runEFunc func(cmd *cobra.Command, _ []string) error

func WithPrettyLogging(
	f func(cmd *cobra.Command, args []string) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := SetupZerolog(cmd)
		if err != nil {
			return err
		}

		return f(cmd, args)
	}
}

func SetupZerolog(cmd *cobra.Command) error {
	var levelStr string
	if v := os.Getenv("EVM_DEBUG"); v != "" {
		levelStr = "debug"
	} else if v := os.Getenv("EVM_LOG_LEVEL"); v != "" {
		levelStr = v
	}

	var out io.Writer = os.Stderr

	if cmd != nil {
		out = cmd.OutOrStderr()
		fl := cmd.Flag("log-level")
		if fl != nil && (fl.Changed || levelStr == "") {
			levelStr = fl.Value.String()
		}
	}

	if levelStr == "" {
		levelStr = "info"
	}

	level, err := zerolog.ParseLevel(levelStr)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = ""

	output := zerolog.ConsoleWriter{Out: out}
	output.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	return nil
}

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
