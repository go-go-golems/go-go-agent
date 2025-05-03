package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/geppetto/pkg/embeddings"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/go-go-agent/pkg/eventbus"
	events "github.com/go-go-golems/go-go-agent/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type LLM interface {
	Generate(ctx context.Context, messages []*conversation.Message) (*conversation.Message, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// WithEventBus sets the EventBus for publishing structured agent events.
func WithEventBus(eb *eventbus.EventBus) GeppettoLLMOption {
	return func(g *GeppettoLLM) error {
		g.eventBus = eb
		return nil
	}
}

// WithRunID sets the Run ID for associating events.
func WithRunID(runID string) GeppettoLLMOption {
	return func(g *GeppettoLLM) error {
		g.runID = &runID
		return nil
	}
}

// NewGeppettoLLM creates a new GeppettoLLM instance.
func NewGeppettoLLM(settings *settings.StepSettings, options ...GeppettoLLMOption) (*GeppettoLLM, error) {
	if settings == nil {
		return nil, errors.New("settings cannot be nil")
	}

	stepSettings := settings.Clone()
	stepSettings.Chat.Stream = true

	embeddingFactory := embeddings.NewSettingsFactoryFromStepSettings(stepSettings)
	embeddingProvider, err := embeddingFactory.NewProvider()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create embedding provider from step settings (type: %s, engine: %s)",
			stepSettings.Embeddings.Type, stepSettings.Embeddings.Engine)
	}

	s := &GeppettoLLM{
		stepSettings:      stepSettings,
		embeddingProvider: embeddingProvider,
	}

	for _, option := range options {
		err := option(s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}

	return s, nil
}

// convertToEventsMessages converts []*Message to []*events.LlmMessage for event logging.
func convertToEventsMessages(msgs []*conversation.Message) []*events.LlmMessage {
	res := make([]*events.LlmMessage, len(msgs))
	for i, m := range msgs {
		// TODO(manuel): Handle multi-content messages for logging?
		// For now, just concatenate text content.
		content := ""
		if m.Content.ContentType() == conversation.ContentTypeChatMessage {
			content += m.Content.String() + "\\n"
		}
		res[i] = &events.LlmMessage{
			Role:    string(m.Content.(*conversation.ChatMessageContent).Role),
			Content: content,
		}
	}
	return res
}

// Chat sends messages to the LLM and returns the response.
// It also emits LLM call events if an EventBus is configured.
func (g *GeppettoLLM) Generate(ctx context.Context, messages []*conversation.Message) (*conversation.Message, error) {
	callID := uuid.New().String() // Unique ID for this specific call

	// --- Emit LlmCallStarted event ---
	if g.eventBus != nil {
		promptPreview := ""
		if len(messages) > 0 {
			// Simple preview of the last message content
			lastMsgContent := messages[len(messages)-1].Content.String()
			maxLen := 100
			if len(lastMsgContent) > maxLen {
				promptPreview = lastMsgContent[:maxLen] + "..."
			} else {
				promptPreview = lastMsgContent
			}
		}

		startPayload := &events.LlmCallStartedPayload{
			AgentClass:    "GeppettoLLM", // Or derive from context if possible later
			Model:         fmt.Sprintf("%s/%v", *g.stepSettings.Chat.ApiType, *g.stepSettings.Chat.Engine),
			Prompt:        convertToEventsMessages(messages), // Use converted []*events.LlmMessage
			PromptPreview: promptPreview,
			CallId:        callID,
			// Step, NodeId, ActionName might be added from context later
		}
		err := g.eventBus.EmitLlmCallStarted(ctx, startPayload, g.runID)
		if err != nil {
			// Log error but continue execution
			log.Warn().Err(err).Msg("Failed to emit LlmCallStarted event")
		}
	}

	chatStep, err := createChatStep(g.stepSettings, g.publisher, g.topicID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chat step")
	}

	// --- Start LLM Call ---
	startTime := time.Now()
	result, err := chatStep.Start(ctx, messages)
	duration := time.Since(startTime)

	results := result.Return()

	// --- Emit LlmCallCompleted event ---
	if g.eventBus != nil {
		var responseStr string
		var resultSummary string
		var errorStr *string
		var tokenUsage *events.TokenUsage

		if err != nil {
			s := "Error: " + err.Error()
			errorStr = &s
			resultSummary = s
		} else if len(results) > 0 {
			responseMsg, err := results[len(results)-1].Value()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get result value")
			}
			responseStr = responseMsg.Content.String()
			resultSummary = summarizeResult(responseStr)
			// TODO(manuel): Extract token usage if available from Geppetto result metadata
			// meta := result.Metadata()
			// if meta != nil { ... }
		} else {
			s := "Received nil result and nil error"
			errorStr = &s
			resultSummary = s
		}

		completePayload := &events.LlmCallCompletedPayload{
			AgentClass:      "GeppettoLLM",
			Model:           fmt.Sprintf("%s/%v", *g.stepSettings.Chat.ApiType, *g.stepSettings.Chat.Engine),
			DurationSeconds: duration.Seconds(),
			Response:        responseStr,
			ResultSummary:   resultSummary,
			Error:           errorStr,
			TokenUsage:      tokenUsage,
			CallId:          callID,
			// Step, NodeId, ActionName might be added from context later
		}
		errEmit := g.eventBus.EmitLlmCallCompleted(ctx, completePayload, g.runID)
		if errEmit != nil {
			// Log error but don't overwrite original execution error
			log.Warn().Err(errEmit).Msg("Failed to emit LlmCallCompleted event")
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to generate")
	}

	// --- Handle LLM Call Result ---
	// Need to handle the case where the stream finishes (err == helpers.ErrCannotReadStreamEnd)
	// but we still got a final message in result.
	if result == nil {
		// This case should ideally not happen if err is nil or ErrCannotReadStreamEnd
		return nil, errors.New("geppetto step returned nil result and nil error unexpectedly")
	}

	if len(results) == 0 {
		return nil, errors.New("no results returned")
	}

	responseMsg, err := results[len(results)-1].Value()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get final result value")
	}

	// Convert the final result message
	return responseMsg, nil
}

// summarizeResult creates a short summary of the result string.
func summarizeResult(resultStr string) string {
	maxLen := 100
	if len(resultStr) > maxLen {
		return resultStr[:maxLen] + "..."
	}
	return resultStr
}
