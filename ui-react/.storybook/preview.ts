import type { Preview } from "@storybook/react";
import 'bootstrap/dist/css/bootstrap.min.css';
import '../src/components/styles.css';

const preview: Preview = {
    parameters: {
        actions: { argTypesRegex: "^on[A-Z].*" },
        controls: {
            matchers: {
                color: /(background|color)$/i,
                date: /Date$/i,
            },
        },
        layout: 'fullscreen',
    },
};

export default preview; 