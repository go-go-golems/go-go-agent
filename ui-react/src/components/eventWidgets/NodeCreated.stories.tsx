import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { NodeCreatedSummary, NodeCreatedTable } from './NodeCreated';

const meta = {
  title: 'EventWidgets/Node/NodeCreated',
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
type SummaryStory = StoryObj<typeof NodeCreatedSummary>;

export const SummaryTaskNode: SummaryStory = {
  render: (args) => <NodeCreatedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-001',
      timestamp: new Date().toISOString(),
      event_type: 'node_created',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        node_nid: '1',
        node_type: 'task',
        task_type: 'research',
        task_goal: 'Find information about the user request and synthesize a comprehensive answer',
        layer: 0,
        root_node_id: 'node-abc-100',
        initial_parent_nids: [],
        step: 1
      }
    }
  },
};

export const SummaryToolNode: SummaryStory = {
  render: (args) => <NodeCreatedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-002',
      timestamp: new Date().toISOString(),
      event_type: 'node_created',
      run_id: 'run-001',
      payload: {
        node_id: 'node-def-456',
        node_nid: '2',
        node_type: 'tool',
        task_type: 'web_search',
        task_goal: 'Search the web for information about machine learning',
        layer: 1,
        root_node_id: 'node-abc-100',
        outer_node_id: 'node-abc-123',
        initial_parent_nids: ['1'],
        step: 2
      }
    }
  },
};

export const SummaryAgentNode: SummaryStory = {
  render: (args) => <NodeCreatedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-003',
      timestamp: new Date().toISOString(),
      event_type: 'node_created',
      run_id: 'run-001',
      payload: {
        node_id: 'node-ghi-789',
        node_nid: '3',
        node_type: 'agent',
        task_type: 'expert',
        task_goal: 'Analyze the research data and provide expert insights',
        layer: 1,
        root_node_id: 'node-abc-100',
        initial_parent_nids: ['1'],
        step: 3
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof NodeCreatedTable>;

export const TableTaskNode: TableStory = {
  render: (args) => <NodeCreatedTable {...args} />,
  args: {
    event: SummaryTaskNode.args?.event,
    compact: true
  },
};

export const TableToolNode: TableStory = {
  render: (args) => <NodeCreatedTable {...args} />,
  args: {
    event: SummaryToolNode.args?.event,
    compact: true
  },
};

export const TableAgentNode: TableStory = {
  render: (args) => <NodeCreatedTable {...args} />,
  args: {
    event: SummaryAgentNode.args?.event,
    compact: true
  },
}; 