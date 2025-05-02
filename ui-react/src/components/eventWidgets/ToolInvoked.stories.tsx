import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { ToolInvokedSummary, ToolInvokedTable } from './ToolInvoked';

const meta = {
  title: 'EventWidgets/Tool/ToolInvoked',
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
type SummaryStory = StoryObj<typeof ToolInvokedSummary>;

export const SummaryWebSearch: SummaryStory = {
  render: (args) => <ToolInvokedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-601',
      timestamp: new Date().toISOString(),
      event_type: 'tool_invoked',
      run_id: 'run-001',
      payload: {
        node_id: 'node-abc-123',
        tool_name: 'web_search',
        api_name: 'mento-tools',
        agent_class: 'ResearchAgent',
        args_summary: 'Search for recent AI developments',
        args: {
          query: 'recent developments in artificial intelligence',
          max_results: 5,
          include_domains: ['academic', 'news'],
          exclude_domains: ['social-media']
        },
        tool_call_id: 'call-123456789',
        step: 5
      }
    }
  },
};

export const SummaryDatabaseQuery: SummaryStory = {
  render: (args) => <ToolInvokedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-602',
      timestamp: new Date().toISOString(),
      event_type: 'tool_invoked',
      run_id: 'run-001',
      payload: {
        node_id: 'node-def-456',
        tool_name: 'database_query',
        api_name: 'sql-connector',
        args_summary: 'Query sales data for Q1 2023',
        args: {
          query: 'SELECT product_name, SUM(sales_amount) as total_sales FROM sales WHERE quarter=1 AND year=2023 GROUP BY product_name ORDER BY total_sales DESC',
          database: 'sales_analytics',
          timeout_seconds: 30
        },
        tool_call_id: 'call-987654321',
        step: 6
      }
    }
  },
};

export const SummaryMinimal: SummaryStory = {
  render: (args) => <ToolInvokedSummary {...args} />,
  args: {
    event: {
      event_id: 'evt-603',
      timestamp: new Date().toISOString(),
      event_type: 'tool_invoked',
      run_id: 'run-001',
      payload: {
        node_id: 'node-ghi-789',
        tool_name: 'simple_calculator',
        args: {
          operation: 'add',
          numbers: [1, 2, 3, 4, 5]
        },
        step: 7
      }
    }
  },
};

// Table Widget Stories
type TableStory = StoryObj<typeof ToolInvokedTable>;

export const TableWebSearch: TableStory = {
  render: (args) => <ToolInvokedTable {...args} />,
  args: {
    event: SummaryWebSearch.args?.event,
    compact: true
  },
};

export const TableDatabaseQuery: TableStory = {
  render: (args) => <ToolInvokedTable {...args} />,
  args: {
    event: SummaryDatabaseQuery.args?.event,
    compact: true
  },
};

export const TableMinimal: TableStory = {
  render: (args) => <ToolInvokedTable {...args} />,
  args: {
    event: SummaryMinimal.args?.event,
    compact: true
  },
}; 