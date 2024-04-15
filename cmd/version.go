package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *CommandHandlers) VersionHandler(cmd *cobra.Command, args []string) {
	// TODO: create constant package
	fmt.Println("gaze-network v0.0.1")
}
