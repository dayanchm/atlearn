package cmd

import (
	"fmt"

	"github.com/dayanchm/at-inspect/internal/atproto"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <handle>",
	Short: "Bir handle'ı DID'e çözer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		did, err := atproto.ResolveHandle(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		fmt.Println(did)
		return nil
	},
}
