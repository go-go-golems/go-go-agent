import type { Preview } from '@storybook/react'
import 'bootstrap/dist/css/bootstrap.min.css'
import '../src/components/styles.css'
import { initialize, mswLoader } from 'msw-storybook-addon'

// Initialize MSW
initialize()

const preview: Preview = {
  loaders: [mswLoader],
  parameters: {
    controls: {
      matchers: {
       color: /(background|color)$/i,
       date: /Date$/i,
      },
    },
    layout: 'centered',
  },
};

export default preview;