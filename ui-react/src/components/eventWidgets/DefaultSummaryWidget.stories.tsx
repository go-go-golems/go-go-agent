import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import DefaultSummaryWidget from './DefaultSummaryWidget.tsx';
import { mockAgentEvent } from './mockData';

const meta = {
  title: 'EventWidgets/DefaultSummaryWidget',
  component: DefaultSummaryWidget,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    event: { control: 'object' },
    setActiveTab: { action: 'tabChanged' },
    onNodeClick: { action: 'nodeClicked' },
  },
  args: {
    onNodeClick: fn(),
    setActiveTab: fn(),
  }
} satisfies Meta<typeof DefaultSummaryWidget>;

export default meta;
type Story = StoryObj<typeof meta>;

// Basic example with a simple payload
export const Basic: Story = {
  args: {
    event: mockAgentEvent('unknown_event_type', { key: 'value', number: 42 }),
  },
};

// More complex payload with nested data
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

// Very large payload to test scrolling and performance
export const LargePayload: Story = {
  args: {
    event: mockAgentEvent('large_event', {
      items: Array.from({ length: 100 }, (_, i) => ({
        id: `item-${i}`,
        value: `Value for item ${i}`,
        metadata: {
          created: new Date().toISOString(),
          tags: ['tag1', 'tag2', 'tag3'],
          stats: {
            views: Math.floor(Math.random() * 1000),
            likes: Math.floor(Math.random() * 100),
            shares: Math.floor(Math.random() * 50),
          }
        }
      }))
    }),
  },
}; 