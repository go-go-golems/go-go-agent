import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { PlanReceivedSummary, PlanReceivedTable } from './PlanReceived';

const meta = {
  title: 'EventWidgets/Graph/PlanReceived',
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

// Sample plan data in different formats
const textPlan = `I will accomplish this task through the following steps:
1. Research the current market trends
2. Analyze the data gathered
3. Create visualizations of the findings
4. Prepare a final report with recommendations`;

const structuredPlan = [
  { id: "step-1", description: "Research the current market trends", estimated_duration: "1 hour" },
  { id: "step-2", description: "Analyze the data gathered", estimated_duration: "2 hours" },
  { id: "step-3", description: "Create visualizations of the findings", estimated_duration: "1.5 hours" },
  { id: "step-4", description: "Prepare a final report with recommendations", estimated_duration: "2 hours" }
];

// Summary Widget Stories
type SummaryStory = StoryObj<typeof PlanReceivedSummary>;

export const SummaryTextPlan: SummaryStory = {
  render: (args) => <PlanReceivedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-301',
      timestamp: new Date().toISOString(),
      event_type: 'plan_received',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        task_type: 'research',
        task_goal: 'Create a comprehensive market analysis report',
        raw_plan: textPlan,
        step: 2
      }
    }
  },
};

export const SummaryStructuredPlan: SummaryStory = {
  render: (args) => <PlanReceivedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-302',
      timestamp: new Date().toISOString(),
      event_type: 'plan_received',
      run_id: 'run-001',
      payload: {
        node_id: 'node-def-456',
        task_type: 'analysis',
        task_goal: 'Analyze market trends and provide recommendations',
        raw_plan: structuredPlan,
        step: 3
      }
    }
  },
};

export const SummaryMinimalPlan: SummaryStory = {
  render: (args) => <PlanReceivedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-303',
      timestamp: new Date().toISOString(),
      event_type: 'plan_received',
      run_id: 'run-001',
      payload: {
        node_id: 'node-ghi-789',
        raw_plan: "Simple one-line plan with no additional structure",
        step: 4
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof PlanReceivedTable>;

export const TableTextPlan: TableStory = {
  render: (args) => <PlanReceivedTable {...args} />,
  args: {
    event: SummaryTextPlan.args?.event,
    compact: true
  },
};

export const TableStructuredPlan: TableStory = {
  render: (args) => <PlanReceivedTable {...args} />,
  args: {
    event: SummaryStructuredPlan.args?.event,
    compact: true
  },
};

export const TableMinimalPlan: TableStory = {
  render: (args) => <PlanReceivedTable {...args} />,
  args: {
    event: SummaryMinimalPlan.args?.event,
    compact: true
  },
}; 