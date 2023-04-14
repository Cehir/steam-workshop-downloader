/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/go-playground/validator/v10"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "config commands",
	Long:  `Shows available config commands and their usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			logger.WithError(err).Error("failed to print help")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func loadConfig() {
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.WithError(err).Fatal("failed to unmarshal config")
	}

	if err := cfg.Validate(); err != nil {
		switch err.(type) {
		case validator.ValidationErrors:
			for _, e := range err.(validator.ValidationErrors) {
				logger.WithField("field", e.Namespace()).WithField("rule", e.Translate(trans)).Error("Validation failed")
			}
		default:
			logger.WithError(err).Error("Config validation failed")
		}
		os.Exit(1)
	}

	logger.WithFields(logger.Fields{
		"cmd":  cfg.Steam.Cmd,
		"user": cfg.Steam.Login.Username,
		"apps": cfg.Apps,
	}).Debug("loading config complete")
}
