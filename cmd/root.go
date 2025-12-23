package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "devops-lab",
	Short: "DevOps Lab Platform CLI",
	Long:  "DevOps Lab - local DevOps platform based on k3d + Helm",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
