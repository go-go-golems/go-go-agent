import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { ToolReturnedSummary, ToolReturnedTable } from './ToolReturned';

const meta = {
  title: 'EventWidgets/Tool/ToolReturned',
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
type SummaryStory = StoryObj<typeof ToolReturnedSummary>;

export const SummaryWebSearchSuccess: SummaryStory = {
  render: (args) => <ToolReturnedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-701',
      timestamp: new Date().toISOString(),
      event_type: 'tool_returned',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        tool_name: 'web_search',
        api_name: 'mento-tools',
        agent_class: 'ResearchAgent',
        state: 'success',
        duration_seconds: 1.75,
        result_summary: 'Found 5 results about recent AI developments',
        result: [
          { title: 'Latest GPT-5 announcement reveals groundbreaking capabilities', url: 'https://example.com/ai-news/1' },
          { title: 'Researchers develop new approach to machine learning', url: 'https://example.com/ai-news/2' },
          { title: 'AI advancements in healthcare show promising results', url: 'https://example.com/ai-news/3' },
          { title: 'New benchmarks for state-of-the-art language models', url: 'https://example.com/ai-news/4' },
          { title: 'Understanding the ethical implications of AI development', url: 'https://example.com/ai-news/5' }
        ],
        tool_call_id: 'call-123456789',
        step: 5
      }
    }
  },
};

export const SummaryDatabaseQueryError: SummaryStory = {
  render: (args) => <ToolReturnedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-702',
      timestamp: new Date().toISOString(),
      event_type: 'tool_returned',
      run_id: 'run-001',
      payload: {
        node_id: 'node-def-456',
        tool_name: 'database_query',
        api_name: 'sql-connector',
        state: 'error',
        duration_seconds: 0.35,
        error: 'Syntax error in SQL query: Unexpected token at position 42',
        result_summary: 'Database query failed',
        tool_call_id: 'call-987654321',
        step: 6
      }
    }
  },
};

export const SummaryCalculatorSuccess: SummaryStory = {
  render: (args) => <ToolReturnedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-703',
      timestamp: new Date().toISOString(),
      event_type: 'tool_returned',
      run_id: 'run-001',
      payload: {
        node_id: 'node-ghi-789',
        tool_name: 'simple_calculator',
        state: 'success',
        duration_seconds: 0.02,
        result_summary: 'Calculation completed',
        result: {
          operation: 'add',
          input: [1, 2, 3, 4, 5],
          output: 15
        },
        step: 7
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof ToolReturnedTable>;

export const TableWebSearchSuccess: TableStory = {
  render: (args) => <ToolReturnedTable {...args} />,
  args: {
    event: SummaryWebSearchSuccess.args?.event,
    compact: true
  },
};

export const TableDatabaseQueryError: TableStory = {
  render: (args) => <ToolReturnedTable {...args} />,
  args: {
    event: SummaryDatabaseQueryError.args?.event,
    compact: true
  },
};

export const TableCalculatorSuccess: TableStory = {
  render: (args) => <ToolReturnedTable {...args} />,
  args: {
    event: SummaryCalculatorSuccess.args?.event,
    compact: true
  },
}; 