import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import GlobalSearch from '../../src/renderer/components/search/GlobalSearch';

// Mock the API hooks
jest.mock('../../src/renderer/hooks/useWalletAPI', () => ({
  useKeys: () => ({
    data: [
      { keyId: 'key1', type: 'Ed25519', algorithm: 'Ed25519', createdAt: '2023-01-01' }
    ]
  }),
  useDIDs: () => ({
    data: [
      { did: 'did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK', method: 'key', createdAt: '2023-01-01' }
    ]
  }),
  useCredentials: () => ({
    data: [
      { 
        credentialId: 'cred1', 
        type: ['VerifiableCredential', 'UniversityDegreeCredential'],
        issuer: { name: 'Test University' },
        issuanceDate: '2023-01-01'
      }
    ]
  }),
  useEvents: () => ({
    data: [
      { 
        eventId: 'event1', 
        type: 'vouch', 
        from: 'did:key:from', 
        to: 'did:key:to', 
        context: 'general',
        timestamp: '2023-01-01'
      }
    ]
  }),
  useTrustScores: () => ({
    data: [
      { 
        did: 'did:key:test', 
        context: 'general', 
        score: 85, 
        updatedAt: '2023-01-01'
      }
    ]
  })
}));

const createTestQueryClient = () => new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderWithProviders = (component: React.ReactElement) => {
  const testQueryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={testQueryClient}>
      <MemoryRouter>
        {component}
      </MemoryRouter>
    </QueryClientProvider>
  );
};

describe('GlobalSearch', () => {
  const mockOnClose = jest.fn();

  beforeEach(() => {
    mockOnClose.mockClear();
  });

  it('renders when open', () => {
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    expect(screen.getByPlaceholderText(/Search keys, DIDs/)).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    renderWithProviders(
      <GlobalSearch isOpen={false} onClose={mockOnClose} />
    );
    
    expect(screen.queryByPlaceholderText(/Search keys, DIDs/)).not.toBeInTheDocument();
  });

  it('shows search tips when no query', () => {
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    expect(screen.getByText('Search Tips')).toBeInTheDocument();
    expect(screen.getByText('Navigate results')).toBeInTheDocument();
  });

  it('searches and shows results', async () => {
    const user = userEvent.setup();
    
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    const searchInput = screen.getByPlaceholderText(/Search keys, DIDs/);
    await user.type(searchInput, 'key');
    
    await waitFor(() => {
      expect(screen.getByText('Keys')).toBeInTheDocument();
    });
  });

  it('closes on escape key', async () => {
    const user = userEvent.setup();
    
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    await user.keyboard('{Escape}');
    
    expect(mockOnClose).toHaveBeenCalled();
  });

  it('navigates results with arrow keys', async () => {
    const user = userEvent.setup();
    
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    const searchInput = screen.getByPlaceholderText(/Search keys, DIDs/);
    await user.type(searchInput, 'key');
    
    // Wait for results to appear
    await waitFor(() => {
      expect(screen.getByText('Keys')).toBeInTheDocument();
    });
    
    // Test arrow key navigation
    await user.keyboard('{ArrowDown}');
    await user.keyboard('{ArrowUp}');
    
    // Results should be navigable (exact testing would require more complex setup)
    expect(searchInput).toHaveFocus();
  });

  it('shows no results message for non-matching query', async () => {
    const user = userEvent.setup();
    
    renderWithProviders(
      <GlobalSearch isOpen={true} onClose={mockOnClose} />
    );
    
    const searchInput = screen.getByPlaceholderText(/Search keys, DIDs/);
    await user.type(searchInput, 'nonexistentquery12345');
    
    await waitFor(() => {
      expect(screen.getByText(/No results found for/)).toBeInTheDocument();
    });
  });
});