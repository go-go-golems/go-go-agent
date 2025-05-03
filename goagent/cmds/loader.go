package cmds

import (
	"io"
	"io/fs"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-agent/goagent/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// AgentCommandLoader loads agent commands from YAML files
type AgentCommandLoader struct{}

// YAMLAgentCommand is a struct used for unmarshaling YAML data
type YAMLAgentCommand struct {
	Name         string                            `yaml:"name"`
	Short        string                            `yaml:"short"`
	Long         string                            `yaml:"long,omitempty"`
	Type         string                            `yaml:"type,omitempty"`
	Flags        []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments    []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	AgentType    string                            `yaml:"agent-type"`
	CommandType  string                            `yaml:"command-type"`
	SystemPrompt string                            `yaml:"system-prompt,omitempty"`
	Prompt       string                            `yaml:"prompt,omitempty"`
	Tools        []string                          `yaml:"tools,omitempty"`
	AgentOptions *types.RawNode                    `yaml:"agent-options,omitempty"`

	// XXX - add LLM profiles
}

const (
	AgentCommandLoaderName = "agent"
)

var _ loaders.CommandLoader = (*AgentCommandLoader)(nil)

// IsFileSupported checks if the file is supported by this loader
func (a *AgentCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	// Check if the file has a YAML extension
	if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
		return false
	}

	// Check if it's an agent command file
	return loaders.CheckYamlFileType(f, fileName, "agent")
}

// LoadCommands implements the CommandLoader interface
func (a *AgentCommandLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	r, err := f.Open(entryName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", entryName)
	}
	defer func(r fs.File) {
		_ = r.Close()
	}(r)

	// Add source tracking option
	sourceOption := cmds.WithSource("file:" + entryName)
	allOptions := append(options, sourceOption)

	return loaders.LoadCommandOrAliasFromReader(
		r,
		a.loadAgentCommandFromReader,
		allOptions,
		aliasOptions)
}

// LoadFromYAML loads Agent commands from YAML content with additional options
func LoadFromYAML(b []byte, options ...cmds.CommandDescriptionOption) ([]cmds.Command, error) {
	loader := &AgentCommandLoader{}
	buf := strings.NewReader(string(b))
	return loader.loadAgentCommandFromReader(buf, options, nil)
}

// loadAgentCommandFromReader loads agent commands from a reader
func (a *AgentCommandLoader) loadAgentCommandFromReader(
	r io.Reader,
	options []cmds.CommandDescriptionOption,
	_ []alias.Option,
) ([]cmds.Command, error) {
	var yamlCmd YAMLAgentCommand

	yamlContent, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read YAML content")
	}

	// Unmarshal the YAML content
	err = yaml.Unmarshal(yamlContent, &yamlCmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal YAML")
	}

	// Create command description from YAML
	cmdDescription := cmds.NewCommandDescription(
		yamlCmd.Name,
		cmds.WithShort(yamlCmd.Short),
		cmds.WithLong(yamlCmd.Long),
		cmds.WithFlags(yamlCmd.Flags...),
		cmds.WithArguments(yamlCmd.Arguments...),
		cmds.WithType(AgentCommandLoaderName), // Set a type for multi-loader support
	)

	// Apply additional options
	for _, option := range options {
		option(cmdDescription)
	}

	var agentCmd cmds.Command
	switch yamlCmd.CommandType {
	case "", "writer":
		agentCmd, err = NewWriterAgentCommand(
			cmdDescription,
			WithAgentType(yamlCmd.AgentType),
			WithSystemPrompt(yamlCmd.SystemPrompt),
			WithPrompt(yamlCmd.Prompt),
			WithTools(yamlCmd.Tools),
			WithAgentOptions(yamlCmd.AgentOptions),
		)

		if err != nil {
			return nil, errors.Wrap(err, "failed to create agent command")
		}
	case "glazed":
		agentCmd, err = NewGlazedAgentCommand(
			cmdDescription,
			WithAgentType(yamlCmd.AgentType),
			WithSystemPrompt(yamlCmd.SystemPrompt),
			WithPrompt(yamlCmd.Prompt),
			WithTools(yamlCmd.Tools),
			WithAgentOptions(yamlCmd.AgentOptions),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create agent command")
		}
	default:
		return nil, errors.Errorf("unknown command type: %s", yamlCmd.CommandType)
	}

	return []cmds.Command{agentCmd}, nil
}
