/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kubeconfig string

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "k8u",
	Short: "Kubernetes User Management",
	Long:  `k8u ist ein Tool, um Kubeconfigs für Entwickler zu erstellen`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "Welcome to kusr")
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "config file (default is $HOME/.kube/config)")
}
