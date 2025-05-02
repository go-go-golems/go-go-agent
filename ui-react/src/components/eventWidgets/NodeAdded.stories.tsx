import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { NodeAddedSummary, NodeAddedTable } from './NodeAdded';

const meta = {
  title: 'EventWidgets/Node/NodeAdded',
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
type SummaryStory = StoryObj<typeof NodeAddedSummary>;

export const SummaryWithGoal: SummaryStory = {
  render: (args) => <NodeAddedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-201',
      timestamp: new Date().toISOString(),
      event_type: 'node_added',
      run_id: 'run-001',
      payload: {
        graph_owner_node_id: 'node-root-001',
        added_node_id: 'node-abc-123',
        added_node_nid: '1.1',
        task_type: 'research',
        task_goal: 'Research information about machine learning algorithms',
        step: 3
      }
    }
  },
};

export const SummaryWithoutGoal: SummaryStory = {
  render: (args) => <NodeAddedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-202',
      timestamp: new Date().toISOString(),
      event_type: 'node_added',
      run_id: 'run-001',
      payload: {
        graph_owner_node_id: 'node-root-001',
        added_node_id: 'node-def-456',
        added_node_nid: '1.2',
        task_type: 'tool',
        step: 4
      }
    }
  },
};

export const SummaryMinimal: SummaryStory = {
  render: (args) => <NodeAddedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-203',
      timestamp: new Date().toISOString(),
      event_type: 'node_added',
      run_id: 'run-001',
      payload: {
        graph_owner_node_id: 'node-root-001',
        added_node_id: 'node-ghi-789',
        added_node_nid: '1.3',
        step: 5
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof NodeAddedTable>;

export const TableWithTaskType: TableStory = {
  render: (args) => <NodeAddedTable {...args} />,
  args: {
    event: SummaryWithGoal.args?.event,
    compact: true
  },
};

export const TableMinimal: TableStory = {
  render: (args) => <NodeAddedTable {...args} />,
  args: {
    event: SummaryMinimal.args?.event,
    compact: true
  },
}; 