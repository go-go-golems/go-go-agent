import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import DefaultTableWidget from './DefaultTableWidget';
import { mockAgentEvent } from './mockData';

const meta = {
  title: 'EventWidgets/DefaultTableWidget',
  component: DefaultTableWidget,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    event: { control: 'object' },
    showCallIds: { control: 'boolean' },
    compact: { control: 'boolean' },
    onNodeClick: { action: 'nodeClicked' },
  },
  args: {
    onNodeClick: fn(),
    showCallIds: true,
    compact: true,
  }
} satisfies Meta<typeof DefaultTableWidget>;

export default meta;
type Story = StoryObj<typeof meta>;

// Basic example with a simple payload
export const Basic: Story = {
  args: {
    event: mockAgentEvent('unknown_event_type', { key: 'value', number: 42 }),
  },
};

// More complex payload
export const Complex: Story = {
  args: {
    event: mockAgentEvent('complex_event', {
      nested: { 
        data: { 
          array: [1, 2, 3], 
          object: { a: 1, b: 2 } 
        }
      },
      status: 'complete',
    }),
  },
};

// Long text to test truncation
export const LongText: Story = {
  args: {
    event: mockAgentEvent('text_event', {
      message: 'This is a very long text that should be truncated in the table view to ensure it does not take up too much space. We want to make sure the ellipsis is displayed properly and the full text is available on hover.'
    }),
  },
};

// With compact mode disabled
export const NotCompact: Story = {
  args: {
    event: mockAgentEvent('unknown_event_type', { key: 'value', number: 42 }),
    compact: false,
  },
}; 