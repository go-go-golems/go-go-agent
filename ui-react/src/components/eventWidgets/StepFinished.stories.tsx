import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { StepFinishedSummary, StepFinishedTable } from './StepFinished';

const meta = {
  title: 'EventWidgets/Step/StepFinished',
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
type SummaryStory = StoryObj<typeof StepFinishedSummary>;

export const SummarySuccess: SummaryStory = {
  render: (args) => <StepFinishedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-003',
      timestamp: new Date().toISOString(),
      event_type: 'step_finished',
      run_id: 'run-001',
      payload: {
        step: 1,
        node_id: 'node-abc-123',
        action_name: 'research_information',
        status_after: 'SUCCESS',
        duration_seconds: 2.45
      }
    }
  },
};

export const SummaryFailed: SummaryStory = {
  render: (args) => <StepFinishedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-004',
      timestamp: new Date().toISOString(),
      event_type: 'step_finished',
      run_id: 'run-001',
      payload: {
        step: 2,
        node_id: 'node-abc-456',
        action_name: 'analyze_data',
        status_after: 'FAILED',
        duration_seconds: 1.23
      }
    }
  },
};

export const SummaryInProgress: SummaryStory = {
  render: (args) => <StepFinishedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-005',
      timestamp: new Date().toISOString(),
      event_type: 'step_finished',
      run_id: 'run-001',
      payload: {
        step: 3,
        node_id: 'node-abc-789',
        action_name: 'generate_report',
        status_after: 'IN_PROGRESS',
        duration_seconds: 0.75
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof StepFinishedTable>;

export const TableSuccess: TableStory = {
  render: (args) => <StepFinishedTable {...args} />,
  args: {
    event: SummarySuccess.args?.event,
    compact: true
  },
};

export const TableFailed: TableStory = {
  render: (args) => <StepFinishedTable {...args} />,
  args: {
    event: SummaryFailed.args?.event,
    compact: true
  },
};

export const TableInProgress: TableStory = {
  render: (args) => <StepFinishedTable {...args} />,
  args: {
    event: SummaryInProgress.args?.event,
    compact: true
  },
}; 