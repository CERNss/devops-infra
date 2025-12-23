package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install infrastructure components",
}

func init() {
	rootCmd.AddCommand(installCmd)
}
