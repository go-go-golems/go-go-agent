import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { StepStartedSummary, StepStartedTable } from './StepStarted';

const meta = {
  title: 'EventWidgets/Step/StepStarted',
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
type SummaryStory = StoryObj<typeof StepStartedSummary>;

export const Summary: SummaryStory = {
  render: (args) => <StepStartedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-001',
      timestamp: new Date().toISOString(),
      event_type: 'step_started',
      run_id: 'run-001',
      payload: {
        step: 1,
        node_id: 'node-abc-123',
        root_id: 'node-root-001',
        node_goal: 'Find information about the topic and summarize it concisely'
      }
    }
  },
};

export const SummaryWithLongGoal: SummaryStory = {
  render: (args) => <StepStartedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-002',
      timestamp: new Date().toISOString(),
      event_type: 'step_started',
      run_id: 'run-001',
      payload: {
        step: 2,
        node_id: 'node-abc-456',
        root_id: 'node-root-001',
        node_goal: 'Analyze the provided code snippet and identify potential improvements, bugs, and optimizations. The code is implementing a complex algorithm for data processing with multiple nested loops and conditionals. Consider time complexity, space complexity, readability, and maintainability in your analysis. Provide specific recommendations for each issue found.'
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof StepStartedTable>;

export const Table: TableStory = {
  render: (args) => <StepStartedTable {...args} />,
  args: {
    event: Summary.args?.event,
    compact: true
  },
};

export const TableWithLongGoal: TableStory = {
  render: (args) => <StepStartedTable {...args} />,
  args: {
    event: SummaryWithLongGoal.args?.event,
    compact: true
  },
}; 