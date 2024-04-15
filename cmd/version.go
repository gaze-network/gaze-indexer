package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show indexer-network version",
		RunE:  versionHandler,
	}
}

func versionHandler(cmd *cobra.Command, args []string) error {
	// TODO: create constant package
	fmt.Println("gaze-network v0.0.1")
	return nil
}
