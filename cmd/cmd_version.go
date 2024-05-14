package cmd

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/constants"
	"github.com/gaze-network/indexer-network/modules/runes"
	"github.com/spf13/cobra"
)

var versions = map[string]string{
	"":      constants.Version,
	"runes": runes.Version,
}

type versionCmdOptions struct {
	Modules string
}

func NewVersionCommand() *cobra.Command {
	opts := &versionCmdOptions{}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show indexer-network version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionHandler(opts, cmd, args)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Modules, "module", "", `Show version of a specific module. E.g. "runes"`)

	return cmd
}

func versionHandler(opts *versionCmdOptions, _ *cobra.Command, _ []string) error {
	version, ok := versions[opts.Modules]
	if !ok {
		// fmt.Fprintln(cmd.ErrOrStderr(), "Unknown module")
		return errors.Wrap(errs.Unsupported, "Invalid module name")
	}
	fmt.Println(version)
	return nil
}
