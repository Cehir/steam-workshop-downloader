/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/Cehir/steam-workshop-downloader/pkg/config"
	"github.com/Cehir/steam-workshop-downloader/pkg/steamcmd"
	"github.com/Cehir/steam-workshop-downloader/pkg/translations/en"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download the configured mods",
	Run: func(cmd *cobra.Command, args []string) {
		loadConfig(false)
		c := steamcmd.NewSteamCmd(&cfg)
		err := c.Download()
		if err != nil {
			logger.WithError(err).Fatal("failed to download mods")
		}
		logger.Debug("download complete")
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	err := en.RegisterDefaultTranslations(config.Validator, trans)
	if err != nil {
		logger.WithError(err).Fatal("Failed to register translations")
		return
	}
}
