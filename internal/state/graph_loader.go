package state

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-agent/internal/db"
)

// LoadGraphFromDB loads the graph state from the database into the graph manager
func LoadGraphFromDB(ctx context.Context, logger zerolog.Logger, dbManager *db.DatabaseManager, graphManager *GraphManager) error {
	// Load graph state first (nodes and edges)
	graphData, err := dbManager.GetLatestRunGraph(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load graph state from database")
		return err
	}

	// Convert DB nodes/edges to state manager format
	graphNodes := make([]GraphNode, 0, len(graphData.Nodes))
	graphEdges := make([]GraphEdge, 0, len(graphData.Edges))

	// Convert nodes
	for nodeID, nodeJSON := range graphData.Nodes {
		var node struct {
			ID          string          `json:"id"`
			NID         string          `json:"nid"`
			Type        string          `json:"type"`
			TaskType    string          `json:"task_type"`
			Goal        string          `json:"goal"`
			Status      string          `json:"status"`
			Layer       int             `json:"layer"`
			OuterNodeID string          `json:"outer_node_id"`
			RootNodeID  string          `json:"root_node_id"`
			Result      json.RawMessage `json:"result"`
			Metadata    json.RawMessage `json:"metadata"`
		}
		if err := json.Unmarshal(nodeJSON, &node); err != nil {
			logger.Warn().Err(err).Str("node_id", nodeID).Msg("Failed to unmarshal node JSON")
			continue
		}
		graphNodes = append(graphNodes, GraphNode{
			NodeID:      node.ID,
			NodeNID:     node.NID,
			NodeType:    node.Type,
			TaskType:    node.TaskType,
			TaskGoal:    node.Goal,
			Status:      node.Status,
			Layer:       node.Layer,
			OuterNodeID: node.OuterNodeID,
			RootNodeID:  node.RootNodeID,
			Result:      node.Result,
			Metadata:    node.Metadata,
		})
	}

	// Convert edges
	for _, edgeJSON := range graphData.Edges {
		var edge struct {
			ID        string          `json:"id"`
			ParentID  string          `json:"parent_id"`
			ChildID   string          `json:"child_id"`
			ParentNID string          `json:"parent_nid"`
			ChildNID  string          `json:"child_nid"`
			Metadata  json.RawMessage `json:"metadata"`
		}
		if err := json.Unmarshal(edgeJSON, &edge); err != nil {
			logger.Warn().Err(err).Msg("Failed to unmarshal edge JSON")
			continue
		}
		graphEdges = append(graphEdges, GraphEdge{
			ID:           edge.ID,
			ParentNodeID: edge.ParentID,
			ChildNodeID:  edge.ChildID,
			ParentNID:    edge.ParentNID,
			ChildNID:     edge.ChildNID,
			Metadata:     edge.Metadata,
		})
	}

	// Load state
	graphManager.LoadStateFromDB(graphNodes, graphEdges)
	return nil
}
