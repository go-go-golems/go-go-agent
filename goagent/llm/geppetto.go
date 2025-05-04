package llm

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/geppetto/pkg/steps/ai"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/chat"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/go-go-agent/pkg/eventbus"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// GeppettoLLM is an implementation of the LLM interface that uses Geppetto for inference and embeddings
type GeppettoLLM struct {
	// stepSettings contains configuration for the LLM
	stepSettings *settings.StepSettings
	// embeddingProvider is used to generate embeddings
	embeddingProvider embeddings.Provider
	// publisher is used to publish events if configured
	publisher message.Publisher
	// topicID is the topic to publish events to if configured
	topicID string
	// Optional event bus for structured events
	eventBus *eventbus.EventBus
	runID    *string // To associate events with a specific run
}

// GeppettoLLMOption defines a function type for configuring GeppettoLLM.
type GeppettoLLMOption func(*GeppettoLLM) error

// WithPublisherAndTopic adds a publisher and topic for event emission
func WithPublisherAndTopic(publisher message.Publisher, topicID string) GeppettoLLMOption {
	return func(llm *GeppettoLLM) error {
		if publisher == nil {
			return errors.New("publisher cannot be nil")
		}
		if topicID == "" {
			return errors.New("topicID cannot be empty")
		}
		llm.publisher = publisher
		llm.topicID = topicID
		return nil
	}
}

// createChatStep creates a new chat step, potentially configured to publish events.
func createChatStep(
	stepSettings *settings.StepSettings,
	publisher message.Publisher,
	topicID string,
) (chat.Step, error) {
	factory := &ai.StandardStepFactory{
		Settings: stepSettings,
	}

	// Configure the underlying chat step
	var stepOptions []chat.StepOption

	// If a publisher is provided (for older streaming), add it to the step options
	if publisher != nil && topicID != "" {
		log.Info().Str("topic", topicID).Msg("Configuring GeppettoLLM with Watermill publisher for streaming")
		stepOptions = append(stepOptions,
			chat.WithPublishedTopic(publisher, topicID),
		)
	}

	step, err := factory.NewStep(stepOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chat step in factory")
	}

	// Add publisher if provided
	if publisher != nil && topicID != "" {
		log.Debug().Str("topic", topicID).Msg("Adding published topic to step")
		err = step.AddPublishedTopic(publisher, topicID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add published topic %s", topicID)
		}
	} else {
		log.Debug().Msg("No publisher or topic ID provided, skipping event publishing setup for step")
	}

	// Type assertion to the specific expected step type
	return step, nil
}

// GenerateEmbedding implements the LLM interface's GenerateEmbedding method
func (g *GeppettoLLM) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if g.embeddingProvider == nil {
		return nil, errors.New("embedding provider is not initialized")
	}
	embedding, err := g.embeddingProvider.GenerateEmbedding(ctx, text)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate embedding")
	}

	return embedding, nil
}

var _ LLM = (*GeppettoLLM)(nil)
