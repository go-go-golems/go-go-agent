package state

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-agent/internal/db"
	"github.com/go-go-golems/go-go-agent/pkg/model"
)

// LoadEventsFromDB loads the event state from the database into the event manager
func LoadEventsFromDB(ctx context.Context, logger zerolog.Logger, dbManager *db.DatabaseManager, eventManager *EventManager) error {
	eventData, err := dbManager.GetLatestRunEvents(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load events from database")
		return err
	}

	// Convert DB events to model.Event objects
	events := make([]model.Event, 0, len(eventData.Events))
	for _, eventJSON := range eventData.Events {
		var event model.Event
		if err := json.Unmarshal(eventJSON, &event); err != nil {
			logger.Warn().Err(err).Msg("Failed to unmarshal event JSON")
			continue
		}
		events = append(events, event)
	}
	eventManager.LoadStateFromDB(events)
	return nil
}
