import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import NetworkStatusIndicator from '../../src/renderer/components/network/NetworkStatusIndicator';

// Mock the API hooks
jest.mock('../../src/renderer/hooks/useWalletAPI', () => ({
  useNetworkStatus: () => ({
    data: { connected: true },
    isLoading: false
  }),
  useNetworkHealth: () => ({
    data: { 
      overall: 'healthy',
      averageLatency: 50,
      connectivity: 95,
      performance: 90,
      reliability: 88
    }
  }),
  useNetworkStats: () => ({
    data: { 
      peerCount: 5,
      messagesPerSecond: 10,
      totalBytesIn: 1024,
      totalBytesOut: 2048
    }
  }),
  useLatestCheckpoint: () => ({
    data: { 
      epoch: '12345',
      timestamp: new Date().toISOString()
    }
  })
}));

// Mock notification context
jest.mock('../../src/renderer/contexts/NotificationContext', () => ({
  useNotification: () => ({
    showNotification: jest.fn()
  })
}));

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderWithQueryClient = (component: React.ReactElement) => {
  const testQueryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={testQueryClient}>
      {component}
    </QueryClientProvider>
  );
};

describe('NetworkStatusIndicator', () => {
  it('renders network status indicator', () => {
    renderWithQueryClient(<NetworkStatusIndicator />);
    
    // Should render the status icon
    expect(screen.getByText('●')).toBeInTheDocument();
  });

  it('shows connected status when network is connected', () => {
    renderWithQueryClient(<NetworkStatusIndicator showDetails={true} />);
    
    expect(screen.getByText('Connected')).toBeInTheDocument();
  });

  it('displays tooltip on hover', async () => {
    renderWithQueryClient(<NetworkStatusIndicator />);
    
    const indicator = screen.getByText('●').closest('.network-status-indicator');
    expect(indicator).toBeInTheDocument();
    
    // Test hover behavior would require user-event
  });

  it('handles loading state', () => {
    // Override the mock for loading state
    jest.doMock('../../src/renderer/hooks/useWalletAPI', () => ({
      useNetworkStatus: () => ({
        data: null,
        isLoading: true
      }),
      useNetworkHealth: () => ({ data: null }),
      useNetworkStats: () => ({ data: null }),
      useLatestCheckpoint: () => ({ data: null })
    }));

    renderWithQueryClient(<NetworkStatusIndicator />);
    
    expect(screen.getByText('⟳')).toBeInTheDocument();
  });
});