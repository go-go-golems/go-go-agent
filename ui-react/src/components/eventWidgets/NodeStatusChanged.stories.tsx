import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { NodeStatusChangedSummary, NodeStatusChangedTable } from './NodeStatusChanged';

const meta = {
  title: 'EventWidgets/Node/NodeStatusChanged',
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
type SummaryStory = StoryObj<typeof NodeStatusChangedSummary>;

export const SummaryInProgressToSuccess: SummaryStory = {
  render: (args) => <NodeStatusChangedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-006',
      timestamp: new Date().toISOString(),
      event_type: 'node_status_changed',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        node_goal: 'Find information about the topic and summarize it concisely',
        old_status: 'IN_PROGRESS',
        new_status: 'SUCCESS',
        step: 1
      }
    }
  },
};

export const SummaryWaitingToInProgress: SummaryStory = {
  render: (args) => <NodeStatusChangedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-007',
      timestamp: new Date().toISOString(),
      event_type: 'node_status_changed',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-456',
        node_goal: 'Analyze the provided data and create summary statistics',
        old_status: 'WAITING',
        new_status: 'IN_PROGRESS',
        step: 2
      }
    }
  },
};

export const SummaryInProgressToFailed: SummaryStory = {
  render: (args) => <NodeStatusChangedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-008',
      timestamp: new Date().toISOString(),
      event_type: 'node_status_changed',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-789',
        node_goal: 'Generate a visualization of the analyzed data',
        old_status: 'IN_PROGRESS',
        new_status: 'FAILED',
        step: 3
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof NodeStatusChangedTable>;

export const TableInProgressToSuccess: TableStory = {
  render: (args) => <NodeStatusChangedTable {...args} />,
  args: {
    event: SummaryInProgressToSuccess.args?.event,
    compact: true
  },
};

export const TableWaitingToInProgress: TableStory = {
  render: (args) => <NodeStatusChangedTable {...args} />,
  args: {
    event: SummaryWaitingToInProgress.args?.event,
    compact: true
  },
};

export const TableInProgressToFailed: TableStory = {
  render: (args) => <NodeStatusChangedTable {...args} />,
  args: {
    event: SummaryInProgressToFailed.args?.event,
    compact: true
  },
}; 