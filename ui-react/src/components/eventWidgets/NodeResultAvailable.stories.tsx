import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { NodeResultAvailableSummary, NodeResultAvailableTable } from './NodeResultAvailable';

const meta = {
  title: 'EventWidgets/Node/NodeResultAvailable',
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
type SummaryStory = StoryObj<typeof NodeResultAvailableSummary>;

export const SummaryTextResult: SummaryStory = {
  render: (args) => <NodeResultAvailableSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-101',
      timestamp: new Date().toISOString(),
      event_type: 'node_result_available',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        action_name: 'research',
        result_summary: 'Found comprehensive information about the requested topic',
        result: 'Artificial Intelligence (AI) is a rapidly evolving field that focuses on creating machines capable of performing tasks that typically require human intelligence. These tasks include learning, reasoning, problem-solving, perception, and language understanding.\n\nKey areas of AI include:\n\n1. Machine Learning: Systems that can learn from data\n2. Natural Language Processing: Understanding and generating human language\n3. Computer Vision: Interpreting and understanding visual information\n4. Robotics: Physical machines capable of interacting with the world',
        step: 5
      }
    }
  },
};

export const SummaryJsonResult: SummaryStory = {
  render: (args) => <NodeResultAvailableSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-102',
      timestamp: new Date().toISOString(),
      event_type: 'node_result_available',
      run_id: 'run-001',
      payload: {
        node_id: 'node-def-456',
        action_name: 'data_analysis',
        result_summary: 'Analyzed the data and produced statistical results',
        result: {
          count: 150,
          mean: 42.5,
          median: 41.2,
          mode: 40,
          standardDeviation: 3.2,
          percentiles: {
            "25": 39.7,
            "50": 41.2,
            "75": 44.8,
            "90": 47.1
          },
          categories: [
            { name: "Category A", value: 45 },
            { name: "Category B", value: 72 },
            { name: "Category C", value: 33 }
          ]
        },
        step: 6
      }
    }
  },
};

export const SummaryNoResult: SummaryStory = {
  render: (args) => <NodeResultAvailableSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-103',
      timestamp: new Date().toISOString(),
      event_type: 'node_result_available',
      run_id: 'run-001',
      payload: {
        node_id: 'node-ghi-789',
        action_name: 'summarize',
        result_summary: 'Summary operation completed',
        result: null,
        step: 7
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof NodeResultAvailableTable>;

export const TableTextResult: TableStory = {
  render: (args) => <NodeResultAvailableTable {...args} />,
  args: {
    event: SummaryTextResult.args?.event,
    compact: true
  },
};

export const TableJsonResult: TableStory = {
  render: (args) => <NodeResultAvailableTable {...args} />,
  args: {
    event: SummaryJsonResult.args?.event,
    compact: true
  },
};

export const TableNoResult: TableStory = {
  render: (args) => <NodeResultAvailableTable {...args} />,
  args: {
    event: SummaryNoResult.args?.event,
    compact: true
  },
}; 