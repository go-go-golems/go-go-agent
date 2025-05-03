package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/clay/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	rootCmd := &cobra.Command{
		Use:   "goagent",
		Short: "GoAgent CLI - Execute AI agents from the command line",
		Long: `GoAgent CLI is a tool for executing AI agents from the command line.
It supports various agent types like ReAct, Plan-and-Execute, and File Collection.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Setup debug logging
			return logging.InitLoggerFromViper()
		},
	}

	// Initialize Viper right after creating the rootCmd
	err = clay.InitViper("pinocchio", rootCmd) // Use "goagent" or a suitable config name
	cobra.CheckErr(err)

	// Setup zerolog
	err = logging.InitLoggerFromViper()
	cobra.CheckErr(err)

	// Setup help system
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Load repository commands
	repoPath := "/home/manuel/code/wesen/corporate-headquarters/go-go-agent/goagent/examples/commands"
	LoadRepositoryCommands(repoPath, rootCmd, helpSystem)

	log.Info().Msg("Starting GoAgent CLI")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
