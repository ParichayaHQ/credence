import React from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { App } from './App';
import { WalletProvider } from './contexts/WalletContext';
import { NotificationProvider } from './contexts/NotificationContext';
import './styles/global.css';

// Simple test to see if React can render at all
console.log('🔍 DEBUG: index.tsx starting with empty publicPath...');

const container = document.getElementById('root');
if (!container) {
  console.error('❌ Root container not found');
  throw new Error('Root container not found');
}

console.log('✅ Root container found:', container);

try {
  const root = createRoot(container);
  console.log('✅ React root created');
  
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: 2,
        staleTime: 5 * 60 * 1000, // 5 minutes
      },
    },
  });

  root.render(
    <React.StrictMode>
      <BrowserRouter>
        <QueryClientProvider client={queryClient}>
          <WalletProvider>
            <NotificationProvider>
              <App />
              <ReactQueryDevtools initialIsOpen={false} />
            </NotificationProvider>
          </WalletProvider>
        </QueryClientProvider>
      </BrowserRouter>
    </React.StrictMode>
  );
  console.log('✅ React render called');
} catch (error) {
  console.error('❌ Error during React setup:', error);
}