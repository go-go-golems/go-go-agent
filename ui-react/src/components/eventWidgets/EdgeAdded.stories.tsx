import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { EdgeAddedSummary, EdgeAddedTable } from './EdgeAdded';

const meta = {
  title: 'EventWidgets/Graph/EdgeAdded',
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    event: { control: 'object' },
    onNodeClick: { action: 'nodeClicked' },
    setActiveTab: { action: 'tabChanged' },
  },
  args: {
    onNodeClick: fn(),
    setActiveTab: fn(),
  }
} satisfies Meta;

export default meta;

// Summary Widget Stories
type SummaryStory = StoryObj<typeof EdgeAddedSummary>;

export const SummaryWithTaskInfo: SummaryStory = {
  render: (args) => <EdgeAddedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-401',
      timestamp: new Date().toISOString(),
      event_type: 'edge_added',
      run_id: 'run-001',
      payload: {
        graph_owner_node_id: 'node-root-001',
        parent_node_id: 'node-abc-123',
        child_node_id: 'node-def-456',
        parent_node_nid: '1',
        child_node_nid: '1.2',
        task_type: 'subtask',
        task_goal: 'Research the specific market segment for the analysis report',
        step: 4
      }
    }
  },
};

export const SummaryMinimal: SummaryStory = {
  render: (args) => <EdgeAddedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-402',
      timestamp: new Date().toISOString(),
      event_type: 'edge_added',
      run_id: 'run-001',
      payload: {
        graph_owner_node_id: 'node-root-001',
        parent_node_id: 'node-ghi-789',
        child_node_id: 'node-jkl-012',
        parent_node_nid: '2',
        child_node_nid: '2.1',
        step: 5
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof EdgeAddedTable>;

export const TableWithTaskInfo: TableStory = {
  render: (args) => <EdgeAddedTable {...args} />,
  args: {
    event: SummaryWithTaskInfo.args?.event,
    compact: true
  },
};

export const TableMinimal: TableStory = {
  render: (args) => <EdgeAddedTable {...args} />,
  args: {
    event: SummaryMinimal.args?.event,
    compact: true
  },
}; 