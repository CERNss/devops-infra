package cmd

import (
	"devops-infra/internal/executor"
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

var execOpts = executor.Options{}

func init() {
	rootCmd.PersistentFlags().BoolVar(&execOpts.Sudo, "sudo", true, "use sudo")
	rootCmd.PersistentFlags().BoolVar(&execOpts.DryRun, "dry-run", false, "dry run")
	rootCmd.PersistentFlags().BoolVar(&execOpts.Verbose, "verbose", false, "verbose output")
}
