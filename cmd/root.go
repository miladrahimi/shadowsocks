package cmd

import (
	"fmt"
	"github.com/miladrahimi/shadowsocks/internal/config"
	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use: "shadowsocks",
}

func init() {
	cobra.OnInitialize(func() { fmt.Println(config.AppName) })

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
