import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import DefaultTableWidget from './DefaultTableWidget';
import { mockAgentEvent, mockStepStartedEvent, mockLlmCompletedEvent } from './mockData';

const meta = {
  title: 'EventWidgets/DefaultTableWidget',
  component: DefaultTableWidget,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    event: { control: 'object' },
    showCallIds: { control: 'boolean' },
    compact: { control: 'boolean' },
    onNodeClick: { action: 'nodeClicked' },
  },
  args: {
    onNodeClick: fn(),
  },
} satisfies Meta<typeof DefaultTableWidget>;

export default meta;
type Story = StoryObj<typeof meta>;

// Basic example with a generic event
export const GenericEvent: Story = {
  args: {
    event: mockAgentEvent('custom_event', {
      message: 'This is a custom event',
      status: 'success',
      details: { key: 'value' },
    }),
    showCallIds: true,
    compact: false,
  },
};

// Example with a step started event
export const StepStartedEvent: Story = {
  args: {
    event: mockStepStartedEvent('node-123'),
    showCallIds: true,
    compact: false,
  },
};

// Example with an LLM call completed event
export const LlmCallCompletedEvent: Story = {
  args: {
    event: mockLlmCompletedEvent('node-456'),
    showCallIds: true,
    compact: false,
  },
};

// Example with compact mode enabled
export const CompactView: Story = {
  args: {
    event: mockStepStartedEvent('node-789'),
    showCallIds: false,
    compact: true,
  },
};

// Example with a long payload to test overflow handling
export const LongPayload: Story = {
  args: {
    event: mockAgentEvent('event_with_long_data', {
      description: 'This is a very long piece of text designed to test how the default widget handles potential overflow or truncation in the table view. It should not break the layout.',
      additionalInfo: {
        field1: 'More data here',
        field2: 'Even more data',
        field3: Array(50).fill('test').join(', '),
      },
    }),
    showCallIds: true,
    compact: false,
  },
}; 