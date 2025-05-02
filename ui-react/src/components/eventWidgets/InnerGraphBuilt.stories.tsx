import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { InnerGraphBuiltSummary, InnerGraphBuiltTable } from './InnerGraphBuilt';

const meta = {
  title: 'EventWidgets/Graph/InnerGraphBuilt',
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
type SummaryStory = StoryObj<typeof InnerGraphBuiltSummary>;

export const SummaryWithTaskInfo: SummaryStory = {
  render: (args) => <InnerGraphBuiltSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-501',
      timestamp: new Date().toISOString(),
      event_type: 'inner_graph_built',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        node_count: 5,
        edge_count: 4,
        node_ids: ['node-def-456', 'node-ghi-789', 'node-jkl-012', 'node-mno-345', 'node-pqr-678'],
        task_type: 'complex_task',
        task_goal: 'Analyze market data and create a comprehensive report with visualizations',
        step: 6
      }
    }
  },
};

export const SummaryMinimal: SummaryStory = {
  render: (args) => <InnerGraphBuiltSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-502',
      timestamp: new Date().toISOString(),
      event_type: 'inner_graph_built',
      run_id: 'run-001',
      payload: {
        node_id: 'node-stu-901',
        node_count: 3,
        edge_count: 2,
        node_ids: ['node-vwx-234', 'node-yza-567', 'node-bcd-890'],
        step: 7
      }
    }
  },
};

export const SummaryEmpty: SummaryStory = {
  render: (args) => <InnerGraphBuiltSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-503',
      timestamp: new Date().toISOString(),
      event_type: 'inner_graph_built',
      run_id: 'run-001',
      payload: {
        node_id: 'node-efg-123',
        node_count: 0,
        edge_count: 0,
        node_ids: [],
        task_type: 'empty_task',
        step: 8
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof InnerGraphBuiltTable>;

export const TableWithTaskType: TableStory = {
  render: (args) => <InnerGraphBuiltTable {...args} />,
  args: {
    event: SummaryWithTaskInfo.args?.event,
    compact: true
  },
};

export const TableMinimal: TableStory = {
  render: (args) => <InnerGraphBuiltTable {...args} />,
  args: {
    event: SummaryMinimal.args?.event,
    compact: true
  },
};

export const TableEmpty: TableStory = {
  render: (args) => <InnerGraphBuiltTable {...args} />,
  args: {
    event: SummaryEmpty.args?.event,
    compact: true
  },
}; 