import React, { createContext, useContext, useState, useCallback, useEffect, ReactNode } from 'react';

interface WalletContextState {
  isInitialized: boolean;
  isLoading: boolean;
  isLocked: boolean;
  error: string | null;
  networkStatus: {
    walletService: boolean;
    lastChecked: string | null;
  };
}

interface WalletContextActions {
  initialize: () => Promise<void>;
  checkLockStatus: () => Promise<boolean>;
  lockWallet: (password: string) => Promise<void>;
  unlockWallet: (password: string) => Promise<void>;
  getNetworkStatus: () => Promise<void>;
  clearError: () => void;
}

type WalletContextType = WalletContextState & WalletContextActions;

const WalletContext = createContext<WalletContextType | undefined>(undefined);

interface WalletProviderProps {
  children: ReactNode;
}

export function WalletProvider({ children }: WalletProviderProps): JSX.Element {
  const [state, setState] = useState<WalletContextState>({
    isInitialized: false,
    isLoading: false,
    isLocked: true,
    error: null,
    networkStatus: {
      walletService: false,
      lastChecked: null,
    },
  });

  const initialize = useCallback(async () => {
    if (!window.walletAPI) {
      setState(prev => ({ 
        ...prev, 
        error: 'Wallet API not available. Please restart the application.',
        isLoading: false 
      }));
      return;
    }

    setState(prev => ({ ...prev, isLoading: true, error: null }));

    try {
      // Check network status first
      await getNetworkStatus();
      
      // Check wallet lock status
      const lockStatus = await checkLockStatus();
      
      setState(prev => ({
        ...prev,
        isInitialized: true,
        isLoading: false,
        isLocked: lockStatus,
      }));
    } catch (error) {
      console.error('Failed to initialize wallet:', error);
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Unknown initialization error',
        isLoading: false,
      }));
    }
  }, []);

  const checkLockStatus = useCallback(async (): Promise<boolean> => {
    try {
      const response = await window.walletAPI.system.checkLockStatus();
      const isLocked = response?.locked ?? true;
      
      setState(prev => ({ ...prev, isLocked }));
      return isLocked;
    } catch (error) {
      console.error('Failed to check lock status:', error);
      // Assume locked on error
      setState(prev => ({ ...prev, isLocked: true }));
      return true;
    }
  }, []);

  const lockWallet = useCallback(async (password: string) => {
    try {
      await window.walletAPI.system.lock(password);
      setState(prev => ({ ...prev, isLocked: true }));
    } catch (error) {
      console.error('Failed to lock wallet:', error);
      throw error;
    }
  }, []);

  const unlockWallet = useCallback(async (password: string) => {
    try {
      await window.walletAPI.system.unlock(password);
      setState(prev => ({ ...prev, isLocked: false }));
    } catch (error) {
      console.error('Failed to unlock wallet:', error);
      throw error;
    }
  }, []);

  const getNetworkStatus = useCallback(async () => {
    try {
      const response = await window.walletAPI.system.getNetworkStatus();
      setState(prev => ({
        ...prev,
        networkStatus: {
          walletService: response?.walletService?.connected ?? false,
          lastChecked: new Date().toISOString(),
        },
      }));
    } catch (error) {
      console.error('Failed to get network status:', error);
      setState(prev => ({
        ...prev,
        networkStatus: {
          walletService: false,
          lastChecked: new Date().toISOString(),
        },
      }));
    }
  }, []);

  const clearError = useCallback(() => {
    setState(prev => ({ ...prev, error: null }));
  }, []);

  // Periodic network status check
  useEffect(() => {
    if (state.isInitialized) {
      const interval = setInterval(getNetworkStatus, 30000); // Every 30 seconds
      return () => clearInterval(interval);
    }
  }, [state.isInitialized, getNetworkStatus]);

  // Listen for wallet status changes from main process
  useEffect(() => {
    if (window.eventAPI) {
      const unsubscribe = window.eventAPI.onWalletStatusChange((locked) => {
        setState(prev => ({ ...prev, isLocked: locked }));
      });

      return unsubscribe;
    }
  }, []);

  const contextValue: WalletContextType = {
    ...state,
    initialize,
    checkLockStatus,
    lockWallet,
    unlockWallet,
    getNetworkStatus,
    clearError,
  };

  return (
    <WalletContext.Provider value={contextValue}>
      {children}
    </WalletContext.Provider>
  );
}

export function useWallet(): WalletContextType {
  const context = useContext(WalletContext);
  if (context === undefined) {
    throw new Error('useWallet must be used within a WalletProvider');
  }
  return context;
}