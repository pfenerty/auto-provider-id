package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var dryRun bool
	var nodeName string

	var rootCmd = &cobra.Command{Use: "auto-provider-id"}

	var cmdMain = &cobra.Command{
		Use:   "operator",
		Short: "Start Operator",
		Run: func(cmd *cobra.Command, args []string) {
			StartOperator(dryRun)
		},
	}

	cmdMain.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "dry run mode")

	var cmdUpdateNode = &cobra.Command{
		Use:   "update-node",
		Short: "Update Node with Provider ID from AWS metadata",
		Run: func(cmd *cobra.Command, args []string) {
			UpdatNode(nodeName, dryRun)
		},
	}

	cmdUpdateNode.Flags().StringVarP(&nodeName, "node", "n", "", "name of the node to label")
	cmdUpdateNode.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "dry run mode")
	cmdUpdateNode.MarkFlagRequired("node")

	rootCmd.AddCommand(cmdMain, cmdUpdateNode)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
