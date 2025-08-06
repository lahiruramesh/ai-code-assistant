package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "dock-route",
	Short: "A CLI tool for managing Docker containers with dynamic subdomains",
	Long: `Docker Route is a CLI tool that helps you deploy and manage
different types of applications (Next.js, React.js, Node.js) using Docker
containers with automatic subdomain routing.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dock-route.yaml)")
	rootCmd.PersistentFlags().StringP("port", "p", "8080", "Port for the reverse proxy server")
	rootCmd.PersistentFlags().StringP("domain", "d", "aicodeagent.abc", "Base domain for subdomains")

	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("domain", rootCmd.PersistentFlags().Lookup("domain"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".dock-route")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
