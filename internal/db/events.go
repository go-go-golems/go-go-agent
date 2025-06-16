package db

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

// GetLatestEvents retrieves the N most recent events across all runs
func (m *DatabaseManager) GetLatestEvents(ctx context.Context, limit int) (*EventData, error) {
	// Get latest events with LIMIT
	rows, err := m.db.QueryContext(ctx, `
		SELECT json_object(
			'event_id', event_id,
			'timestamp', timestamp,
			'event_type', event_type,
			'payload', json(payload),
			'run_id', run_id
		) AS event_json
		FROM events
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query latest events")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.logger.Error().Err(err).Msg("Error closing rows")
		}
	}()

	data := &EventData{
		Events: []json.RawMessage{},
	}

	for rows.Next() {
		var eventJSON []byte
		if err := rows.Scan(&eventJSON); err != nil {
			return nil, errors.Wrap(err, "failed to scan event row")
		}
		data.Events = append(data.Events, eventJSON)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating event rows")
	}

	// Reverse the order to display oldest first
	for i, j := 0, len(data.Events)-1; i < j; i, j = i+1, j-1 {
		data.Events[i], data.Events[j] = data.Events[j], data.Events[i]
	}

	return data, nil
}

// GetEventsInTimeRange retrieves events within a specified time range
func (m *DatabaseManager) GetEventsInTimeRange(ctx context.Context, startTime, endTime string) (*EventData, error) {
	// Query events within time range
	rows, err := m.db.QueryContext(ctx, `
		SELECT json_object(
			'event_id', event_id,
			'timestamp', timestamp,
			'event_type', event_type,
			'payload', json(payload),
			'run_id', run_id
		) AS event_json
		FROM events
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`, startTime, endTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events in time range")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.logger.Error().Err(err).Msg("Error closing rows")
		}
	}()

	data := &EventData{
		Events: []json.RawMessage{},
	}

	for rows.Next() {
		var eventJSON []byte
		if err := rows.Scan(&eventJSON); err != nil {
			return nil, errors.Wrap(err, "failed to scan event row")
		}
		data.Events = append(data.Events, eventJSON)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating event rows")
	}

	return data, nil
}

// GetAvailableRunIDs retrieves a list of all available run IDs
func (m *DatabaseManager) GetAvailableRunIDs(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT run_id
		FROM runs
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query run IDs")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			m.logger.Error().Err(err).Msg("Error closing rows")
		}
	}()

	var runIDs []string
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			return nil, errors.Wrap(err, "failed to scan run ID")
		}
		runIDs = append(runIDs, runID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating run ID rows")
	}

	return runIDs, nil
}
