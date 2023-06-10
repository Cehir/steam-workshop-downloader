/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Cehir/steam-workshop-downloader/pkg/output"

	"github.com/spf13/cobra"
)

// configShow represents the export command
var configShow = &cobra.Command{
	Use:     "show",
	Short:   "print the current config",
	Aliases: []string{"export"},
	Run: func(cmd *cobra.Command, args []string) {
		loadConfig(true)
		_ = out.Print(&cfg)
	},
}

var (
	out = output.YAML
)

func init() {
	configCmd.AddCommand(configShow)

	// Cobra supports local flags
	configShow.Flags().VarP(&out, "out", "o", "output format (yaml or json)")
}
