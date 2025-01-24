/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8u",
	Short: "Kubernetes User Management",
	Long:  `k8u ist ein Tool, um Kubeconfigs für Entwickler zu erstellen`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.medicproof-cli.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".medicproof-cli" (without extension).
		viper.AddConfigPath(home)
		// viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".medicproof-cli")
	}
	viper.SetEnvPrefix("PCCLI")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.GetViper().ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())

}
