import React, { useEffect, useState } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/layout/Layout';
import { Dashboard } from './pages/Dashboard';
import { Keys } from './pages/Keys';
import { DIDs } from './pages/DIDs';
import { Credentials } from './pages/Credentials';
import { TrustScores } from './pages/TrustScores';
import { Events } from './pages/Events';
import { Network } from './pages/Network';
import { Settings } from './pages/Settings';
import { WalletSetup } from './pages/WalletSetup';
import { useWallet } from './contexts/WalletContext';
import { useNotification } from './contexts/NotificationContext';
import { LoadingScreen } from './components/common/LoadingScreen';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import AppLockManager from './components/security/AppLockManager';
import SessionManager from './components/security/SessionManager';
import OnboardingWizard from './components/onboarding/OnboardingWizard';

export function App(): JSX.Element {
  const { isInitialized, isLoading, error, initialize } = useWallet();
  const { showNotification } = useNotification();
  const [appLocked, setAppLocked] = useState(false);
  const [showOnboarding, setShowOnboarding] = useState(false);

  useEffect(() => {
    // Initialize wallet on app start
    initialize();

    // Check if onboarding is needed
    const onboardingCompleted = localStorage.getItem('onboardingCompleted');
    if (!onboardingCompleted) {
      setShowOnboarding(true);
    }
  }, [initialize]);

  useEffect(() => {
    // Listen for wallet service events
    const handleServiceReady = () => {
      showNotification('Wallet service connected', 'success');
    };

    const handleServiceError = (error: { message: string; error: string }) => {
      showNotification(
        `Service Error: ${error.message}`,
        'error',
        error.error
      );
    };

    // Listen for system events (if available)
    if (window.eventAPI) {
      const unsubscribeNetwork = window.eventAPI.onNetworkStatusChange((status) => {
        const isConnected = status?.walletService?.connected;
        showNotification(
          `Network ${isConnected ? 'Connected' : 'Disconnected'}`,
          isConnected ? 'success' : 'warning'
        );
      });

      const unsubscribeWallet = window.eventAPI.onWalletStatusChange((locked) => {
        showNotification(
          `Wallet ${locked ? 'Locked' : 'Unlocked'}`,
          locked ? 'warning' : 'success'
        );
      });

      const unsubscribeEvents = window.eventAPI.onNewEvent((event) => {
        showNotification(
          `New ${event.type}: ${event.from} → ${event.to}`,
          'info'
        );
      });

      return () => {
        unsubscribeNetwork();
        unsubscribeWallet();
        unsubscribeEvents();
      };
    }
  }, [showNotification]);

  // Show loading screen while initializing
  if (isLoading) {
    return <LoadingScreen message="Initializing wallet..." />;
  }

  // Show error screen if initialization failed
  if (error) {
    return (
      <div className="app-error">
        <div className="error-container">
          <h1>Failed to Initialize Wallet</h1>
          <p>{error}</p>
          <button onClick={() => window.location.reload()}>
            Retry
          </button>
        </div>
      </div>
    );
  }

  // Show onboarding if not completed
  if (showOnboarding) {
    return <OnboardingWizard />;
  }

  // Show setup screen if wallet is not initialized
  if (!isInitialized) {
    return <WalletSetup />;
  }

  return (
    <ErrorBoundary>
      <SessionManager
        onSessionExpired={() => {
          showNotification('Session expired. Please restart the application.', 'warning');
        }}
        onSessionWarning={(minutesLeft) => {
          showNotification(`Session expires in ${minutesLeft} minutes`, 'warning');
        }}
      >
        <AppLockManager onLockStateChange={setAppLocked}>
          <div className="app">
            <Layout>
              <Routes>
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/keys" element={<Keys />} />
                <Route path="/dids" element={<DIDs />} />
                <Route path="/credentials" element={<Credentials />} />
                <Route path="/trust-scores" element={<TrustScores />} />
                <Route path="/events" element={<Events />} />
                <Route path="/network" element={<Network />} />
                <Route path="/settings" element={<Settings />} />
              </Routes>
            </Layout>
          </div>
        </AppLockManager>
      </SessionManager>
    </ErrorBoundary>
  );
}