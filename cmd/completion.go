package cmd

import (
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long: `To load completion now, run

. ~/.nostromo/completion

To configure your bash shell to load completions for each session add to your shell init files

# In ~/.bashrc or ~/.profile
. ~/.nostromo/completion`,
	Run: func(cmd *cobra.Command, args []string) {
		err := rootCmd.GenBashCompletionFile(pathutil.Abs("~/.nostromo/completion"))
		if err != nil {
			log.Error(err)
			printUsage(cmd)
			return
		}
		log.Regular(cmd.Long)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
