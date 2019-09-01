package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config item in manifest",
	Long: `Set a config item in manifest.
Nostromo config items are saved in the manifest.

Use this command to set values for these settings:
verbose: true | false`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		task.SetConfig(args[0], args[1])
	},
}

func init() {
	manifestCmd.AddCommand(setCmd)
}
