import type { Meta, StoryObj } from '@storybook/react';
import { fn } from '@storybook/test';
import { LlmCallStartedSummary, LlmCallStartedTable, LlmCallStartedPromptTab } from './LlmCallStarted';
import { mockLlmCallStartedEvent } from './mockData';

// Common story setup for LlmCallStartedSummary
const summaryMeta = {
  title: 'EventWidgets/LlmCall/LlmCallStartedSummary',
  component: LlmCallStartedSummary,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    setActiveTab: { action: 'tabChanged' },
    onNodeClick: { action: 'nodeClicked' },
  },
  args: {
    onNodeClick: fn(),
    setActiveTab: fn(),
    event: mockLlmCallStartedEvent(),
  }
} satisfies Meta<typeof LlmCallStartedSummary>;

export default summaryMeta;
type SummaryStory = StoryObj<typeof summaryMeta>;

// Default summary story
export const Default: SummaryStory = {};

// Custom summary with longer prompt
export const LongPrompt: SummaryStory = {
  args: {
    event: mockLlmCallStartedEvent('node-custom', undefined, {
      prompt: `This is a really long prompt that will be truncated in the preview.
      
It has multiple paragraphs and formatting to test how the component handles long content.

# Example Heading
      
Here's some additional content to make it even longer.
* Point 1
* Point 2
* Point 3

And here's even more content to ensure we hit the truncation limit.`
    }),
  },
};

// Table widget stories
export const Table = {
  title: 'EventWidgets/LlmCall/LlmCallStartedTable',
  component: LlmCallStartedTable,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    compact: { control: 'boolean' },
    showCallIds: { control: 'boolean' },
    onNodeClick: { action: 'nodeClicked' },
  },
  render: (args) => <div style={{ width: '400px', padding: '10px', border: '1px solid #ccc' }}>
    <LlmCallStartedTable {...args} />
  </div>,
} satisfies Meta<typeof LlmCallStartedTable>;

export const TableDefault: StoryObj<typeof Table> = {
  args: {
    event: mockLlmCallStartedEvent(),
    compact: true,
    showCallIds: true,
    onNodeClick: fn(),
  },
};

export const TableNotCompact: StoryObj<typeof Table> = {
  args: {
    event: mockLlmCallStartedEvent(),
    compact: false,
    showCallIds: true,
    onNodeClick: fn(),
  },
};

// Prompt tab stories
export const PromptTab = {
  title: 'EventWidgets/LlmCall/LlmCallStartedPromptTab',
  component: LlmCallStartedPromptTab,
  parameters: { layout: 'centered' },
  tags: ['autodocs'],
  argTypes: {
    onNodeClick: { action: 'nodeClicked' },
  },
} satisfies Meta<typeof LlmCallStartedPromptTab>;

export const PromptTabDefault: StoryObj<typeof PromptTab> = {
  args: {
    event: mockLlmCallStartedEvent(),
    tabKey: 'prompt',
    onNodeClick: fn(),
  },
};

export const PromptTabMarkdown: StoryObj<typeof PromptTab> = {
  args: {
    event: mockLlmCallStartedEvent('node-custom', undefined, {
      prompt: `# Markdown Prompt

This is a prompt with *markdown* formatting to test how the component renders markdown.

## Code Example

\`\`\`python
def hello_world():
    print("Hello, world!")
\`\`\`

- List item 1
- List item 2
- List item 3
`
    }),
    tabKey: 'prompt',
    onNodeClick: fn(),
  },
}; 