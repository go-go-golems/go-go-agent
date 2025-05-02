// RegisterAgentFactories initializes a registry with all available agent factories
func RegisterAgentFactories(registry *AgentRegistry) {
	// Add the ExecutorAgentFactory to the registry
	registry.Register(ExecutorAgentType, &ExecutorAgentFactory{})
	// Add the FileCollectionAgentFactory to the registry
	registry.Register(FileCollectionAgentType, &FileCollectionAgentFactory{})
	// Add the StructuredDataAgentFactory to the registry
	registry.Register(StructuredDataAgentType, &StructuredDataAgentFactory{})
} 