package app

import (
	"fmt"
	"github.com/gostship/kunkka/pkg/version"
	"github.com/spf13/cobra"
	"os"
)

// NewCmdVersion returns a cobra command for fetching versions
func NewCmdVersion(cli *KunkkaCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the server version information",
		Long:  "Print the server version information for the current context",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.GetVersion()
			fmt.Fprintf(os.Stdout, "version: %v\n", v.String())
		},
	}
	return cmd
}
