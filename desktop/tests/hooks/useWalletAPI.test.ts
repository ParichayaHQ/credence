import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useKeys, useGenerateKey } from '../../src/renderer/hooks/useWalletAPI';

// Mock the notification context
jest.mock('../../src/renderer/contexts/NotificationContext', () => ({
  useNotification: () => ({
    showNotification: jest.fn()
  })
}));

// Mock window.walletAPI
const mockWalletAPI = {
  keys: {
    list: jest.fn(),
    get: jest.fn(),
    generate: jest.fn(),
    delete: jest.fn()
  }
};

(global as any).window = {
  ...global.window,
  walletAPI: mockWalletAPI
};

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('useWalletAPI hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useKeys', () => {
    it('fetches keys successfully', async () => {
      const mockKeys = [
        { keyId: 'key1', type: 'Ed25519', algorithm: 'Ed25519' }
      ];
      
      mockWalletAPI.keys.list.mockResolvedValue({ data: mockKeys });

      const { result } = renderHook(() => useKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockKeys);
      });

      expect(mockWalletAPI.keys.list).toHaveBeenCalled();
    });

    it('handles empty keys list', async () => {
      mockWalletAPI.keys.list.mockResolvedValue({ data: [] });

      const { result } = renderHook(() => useKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.data).toEqual([]);
      });
    });

    it('handles keys list error', async () => {
      mockWalletAPI.keys.list.mockRejectedValue(new Error('API Error'));

      const { result } = renderHook(() => useKeys(), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.error).toBeTruthy();
      });
    });
  });

  describe('useGenerateKey', () => {
    it('generates key successfully', async () => {
      const mockNewKey = { keyId: 'newkey', type: 'Ed25519' };
      mockWalletAPI.keys.generate.mockResolvedValue({ data: mockNewKey });

      const { result } = renderHook(() => useGenerateKey(), {
        wrapper: createWrapper(),
      });

      result.current.mutate('Ed25519');

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockWalletAPI.keys.generate).toHaveBeenCalledWith('Ed25519');
    });

    it('handles key generation error', async () => {
      mockWalletAPI.keys.generate.mockRejectedValue(new Error('Generation failed'));

      const { result } = renderHook(() => useGenerateKey(), {
        wrapper: createWrapper(),
      });

      result.current.mutate('Ed25519');

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });
    });
  });
});