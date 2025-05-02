import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { LlmCallCompletedSummary, LlmCallCompletedTable, LlmCallCompletedResponseTab } from './LlmCallCompleted';
import { mockLlmCallCompletedEvent } from './mockData';

// Common story setup for LlmCallCompletedSummary
const summaryMeta = {
  title: 'EventWidgets/LlmCall/LlmCallCompletedSummary',
  component: LlmCallCompletedSummary,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    setActiveTab: { action: 'tabChanged' },
    onNodeClick: { action: 'nodeClicked' },
  },
  args: {
    onNodeClick: fn(),
    setActiveTab: fn(),
    event: mockLlmCallCompletedEvent(),
  }
} satisfies Meta<typeof LlmCallCompletedSummary>;

export default summaryMeta;
type SummaryStory = StoryObj<typeof summaryMeta>;

// Default summary story
export const Default: SummaryStory = {};

// Custom summary with longer response
export const LongResponse: SummaryStory = {
  args: {
    event: mockLlmCallCompletedEvent('node-custom', undefined, {
      response: `This is a really long response that will be truncated in the preview.
      
Artificial Intelligence (AI) refers to the simulation of human intelligence in machines that are programmed to think and learn like humans. The term can also be applied to any machine that exhibits traits associated with a human mind such as learning and problem-solving.

The ideal characteristic of artificial intelligence is its ability to rationalize and take actions that have the best chance of achieving a specific goal. When most people hear the term artificial intelligence, the first thing they usually think of is robots. That's because big-budget films and novels weave stories about human-like machines that wreak havoc on Earth.

But nothing could be further from the truth. Artificial intelligence is based on the principle that human intelligence can be defined in a way that a machine can easily mimic it and execute tasks, from the simplest to those that are even more complex. The goals of artificial intelligence include learning, reasoning, and perception.

As technology advances, previous benchmarks that defined artificial intelligence become outdated. For example, machines that calculate basic functions or recognize text through optical character recognition are no longer considered to embody artificial intelligence, since this function is now taken for granted as an inherent computer function.

AI is continuously evolving to benefit many different industries. Machines are wired using a cross-disciplinary approach based on mathematics, computer science, linguistics, psychology, and more.`
    }),
  },
};

// Table widget stories
export const Table = {
  title: 'EventWidgets/LlmCall/LlmCallCompletedTable',
  component: LlmCallCompletedTable,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    compact: { control: 'boolean' },
    showCallIds: { control: 'boolean' },
    onNodeClick: { action: 'nodeClicked' },
  },
  render: (args) => <div style={{ width: '400px', padding: '10px', border: '1px solid #ccc' }}>
    <LlmCallCompletedTable {...args} />
  </div>,
} satisfies Meta<typeof LlmCallCompletedTable>;

export const TableDefault: StoryObj<typeof Table> = {
  args: {
    event: mockLlmCallCompletedEvent(),
    compact: true,
    showCallIds: true,
    onNodeClick: fn(),
  },
};

export const TableNotCompact: StoryObj<typeof Table> = {
  args: {
    event: mockLlmCallCompletedEvent(),
    compact: false,
    showCallIds: true,
    onNodeClick: fn(),
  },
};

// Response tab stories
export const ResponseTab = {
  title: 'EventWidgets/LlmCall/LlmCallCompletedResponseTab',
  component: LlmCallCompletedResponseTab,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    onNodeClick: { action: 'nodeClicked' },
  },
} satisfies Meta<typeof LlmCallCompletedResponseTab>;

export const ResponseTabDefault: StoryObj<typeof ResponseTab> = {
  args: {
    event: mockLlmCallCompletedEvent(),
    tabKey: 'response',
    onNodeClick: fn(),
  },
};

export const ResponseTabMarkdown: StoryObj<typeof ResponseTab> = {
  args: {
    event: mockLlmCallCompletedEvent('node-custom', undefined, {
      response: `# AI Research Findings

Based on my analysis, here are the key points about artificial intelligence:

## Key Components

1. **Machine Learning** - Algorithms that improve through experience
2. **Neural Networks** - Computing systems inspired by biological neural networks
3. **Deep Learning** - Neural networks with many layers

## Code Example

\`\`\`python
import tensorflow as tf

model = tf.keras.Sequential([
    tf.keras.layers.Dense(128, activation='relu'),
    tf.keras.layers.Dense(10, activation='softmax')
])
\`\`\`

## Current Applications

- Natural language processing
- Computer vision
- Autonomous vehicles
- Healthcare diagnostics
`
    }),
    tabKey: 'response',
    onNodeClick: fn(),
  },
}; 