package main

import (
	"os"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/help"
	goagentcmds "github.com/go-go-golems/go-go-agent/goagent/cmds"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// LoadRepositoryCommands loads agent commands from the specified repository path and adds them to the root command.
func LoadRepositoryCommands(repoPath string, rootCmd *cobra.Command, helpSystem *help.HelpSystem) {
	// check if repoPath exists and is a directory
	if fi, err := os.Stat(repoPath); err == nil && fi.IsDir() {
		loader := &goagentcmds.AgentCommandLoader{}
		repo := repositories.NewRepository(
			repositories.WithDirectories(repositories.Directory{
				FS:            os.DirFS(repoPath),
				RootDirectory: ".",
				Name:          "file-agents",
				SourcePrefix:  "file",
			}),
			repositories.WithCommandLoader(loader),
		)

		// Load commands into the repository
		err = repo.LoadCommands(helpSystem)
		if err != nil {
			log.Warn().Err(err).Str("path", repoPath).Msg("Error loading commands from repository")
			// Don't exit, maybe other commands loaded fine
			return
		}

		// Collect commands from the repository
		loadedCommands := repo.CollectCommands([]string{}, true)
		log.Info().Int("count", len(loadedCommands)).Str("path", repoPath).Msg("Loaded commands from repository")

		// Add commands from repository to Cobra
		for _, cmd := range loadedCommands {
			cobraCmd, err := pinocchio_cmds.BuildCobraCommandWithGeppettoMiddlewares(cmd)
			if err != nil {
				log.Error().Err(err).Str("command", cmd.Description().Name).Msg("Error building cobra command from repository agent command")
				continue
			}
			rootCmd.AddCommand(cobraCmd)
		}
	} else {
		log.Warn().Str("path", repoPath).Msg("Repository path does not exist or is not a directory, skipping.")
	}
}
