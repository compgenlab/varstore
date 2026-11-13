package cli

import "github.com/spf13/cobra"

// InitCmd registers every subcommand with the root. Order here is listing
// order in --help, which follows the pipeline: build, extend, read.
func InitCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(convertCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(refBlockCalcCmd)
	rootCmd.AddCommand(refBlockAttachCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(summaryCmd)
}
