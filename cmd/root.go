/*
Copyright © 2023 André Oehmicke oehmicke@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/Cehir/steam-workshop-downloader/pkg/config"
	english "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"os"
	"strings"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	verbose     bool
	veryVerbose bool
	cfg         config.Config
)

// translations
var (
	eng      = english.New()
	uni      = ut.New(eng, eng)
	trans, _ = uni.GetTranslator("en")
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "steam-workshop-downloader",
	Short: "A client to manage mods from the steam workshop.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initConfig)

	// global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.steam-workshop-downloader.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "info output")
	rootCmd.PersistentFlags().BoolVar(&veryVerbose, "vv", false, "debug output")
}

func initLogger() {
	logger.SetLevel(logger.WarnLevel)
	if verbose {
		logger.SetLevel(logger.InfoLevel)
	}
	if veryVerbose {
		logger.SetLevel(logger.DebugLevel)
	}
	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	logger.Debug("initConfig called")

	viper.SetEnvPrefix("swd")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("steam.cmd", config.DefaultSteamCMDPath())
	viper.SetDefault("steam.login.username", "anonymous")
	viper.SetDefault("steam.login.password", "")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		// Search config in home directory with name ".steam-workshop-downloader" (without extension).
		viper.AddConfigPath(home)

		// Search config in current folder
		dir, err := os.Getwd()
		cobra.CheckErr(err)
		viper.AddConfigPath(dir)

		viper.SetConfigType("yaml")
		viper.SetConfigName(".steam-workshop-downloader")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Debug("Using config file:", viper.ConfigFileUsed())
	}
}
