package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "earhw",
	Short: "Hardware information tool with TUI",
	Long:  "A hardware information tool that displays CPU, RAM, and disk information in a retro-style TUI interface.",
	Run:   runTUI,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
