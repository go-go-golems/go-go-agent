import type { Meta, StoryObj } from '@storybook/react';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { http, HttpResponse } from 'msw'; // Import MSW handlers
import { eventsApi, EventsApiResponse, ConnectionStatus } from '../features/events/eventsApi';
import modalStackReducer from '../features/ui/modalStackSlice';
import EventTable from './EventTable';
import { generateMockEventSequence } from './eventWidgets/mockData';

// Create a real store (MSW will intercept the actual fetch)
const store = configureStore({
    reducer: {
        [eventsApi.reducerPath]: eventsApi.reducer,
        modalStack: modalStackReducer,
    },
    middleware: (getDefaultMiddleware) =>
        getDefaultMiddleware().concat(eventsApi.middleware),
});

// Mock the API response data
const mockEvents = generateMockEventSequence(20);

// Define the success response structure
const successResponse: EventsApiResponse = {
    events: mockEvents,
    status: ConnectionStatus.Connected,
};

// MSW handlers
const handlers = {
    default: [ // Handler for the default success case
        http.get('/api/events', () => {
            return HttpResponse.json(successResponse);
        })
    ],
    loading: [ // Handler for the loading state (might need adjustment based on how loading is triggered)
        http.get('/api/events', async () => {
            // Simulate delay for loading state visualization
            await new Promise(resolve => setTimeout(resolve, 3000));
            return HttpResponse.json(successResponse); // Eventually return success, or handle differently if needed
        })
    ],
    error: [ // Handler for the error case
        http.get('/api/events', () => {
            return new HttpResponse('Internal Server Error', { status: 500 });
        })
    ],
};


const meta = {
    title: 'Components/EventTable',
    component: EventTable,
    // Apply the store provider globally for all stories in this file
    decorators: [
        (Story) => (
            <Provider store={store}>
                <Story />
            </Provider>
        )
    ],
    parameters: {
        layout: 'fullscreen',
        // Define MSW handlers for stories
        msw: {
            handlers: handlers.default // Default to success handlers
        }
    },
    tags: ['autodocs'],
} satisfies Meta<typeof EventTable>;

export default meta;
type Story = StoryObj<typeof meta>;

// Basic story using the default success handlers
export const Default: Story = {};

// Story showing loading state (MSW handler simulates delay)
// Note: The visual loading state depends on the component's implementation.
// This mock only simulates a delayed response.
export const Loading: Story = {
    parameters: {
        msw: {
            handlers: handlers.loading
        }
    }
};

// Story showing error state
export const Error: Story = {
    parameters: {
        msw: {
            handlers: handlers.error
        }
    }
}; 