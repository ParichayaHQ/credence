import React from 'react';
import { useWallet } from '../contexts/WalletContext';
import { useKeys, useDIDs, useCredentials, useEvents } from '../hooks/useWalletAPI';

export function Dashboard(): JSX.Element {
  const { networkStatus, isLocked } = useWallet();
  const { data: keys = [] } = useKeys();
  const { data: dids = [] } = useDIDs();
  const { data: credentials = [] } = useCredentials();
  const { data: events = [] } = useEvents();

  const quickActions = [
    {
      label: 'Generate Key',
      icon: '🔑',
      path: '/keys',
      description: 'Create a new cryptographic key'
    },
    {
      label: 'Create DID',
      icon: '🆔',
      path: '/dids',
      description: 'Establish your digital identity'
    },
    {
      label: 'Import Credential',
      icon: '📜',
      path: '/credentials',
      description: 'Add a verifiable credential'
    },
    {
      label: 'Create Event',
      icon: '📝',
      path: '/events',
      description: 'Create a vouch or report'
    },
  ];

  const navigateTo = (path: string) => {
    window.location.hash = `#${path}`;
  };

  return (
    <div className="page dashboard">
      <div className="page-header">
        <h1 className="page-title">Dashboard</h1>
        <p className="page-subtitle">Overview of your wallet status and recent activity</p>
      </div>

      <div className="dashboard-content">
        <div className="dashboard-grid">
          {/* Status Cards */}
          <div className="status-card">
            <div className="card-header">
              <h3 className="card-title">Wallet Status</h3>
            </div>
            <div className="card-content">
              <div className="status-item">
                <span className="status-label">Lock Status:</span>
                <span className={`status-value ${isLocked ? 'locked' : 'unlocked'}`}>
                  {isLocked ? 'Locked 🔒' : 'Unlocked 🔓'}
                </span>
              </div>
            </div>
          </div>

          <div className="status-card">
            <div className="card-header">
              <h3 className="card-title">Network Status</h3>
            </div>
            <div className="card-content">
              <div className="status-item">
                <span className="status-label">Wallet Service:</span>
                <span className={`status-value ${networkStatus.walletService ? 'connected' : 'disconnected'}`}>
                  {networkStatus.walletService ? 'Connected ✅' : 'Disconnected ❌'}
                </span>
              </div>
              {networkStatus.lastChecked && (
                <div className="status-item">
                  <span className="status-label">Last Checked:</span>
                  <span className="status-value">
                    {new Date(networkStatus.lastChecked).toLocaleString()}
                  </span>
                </div>
              )}
            </div>
          </div>

          <div className="status-card">
            <div className="card-header">
              <h3 className="card-title">Quick Stats</h3>
            </div>
            <div className="card-content">
              <div className="stats-grid">
                <div className="stat-item" onClick={() => navigateTo('/keys')}>
                  <div className="stat-value">{keys.length}</div>
                  <div className="stat-label">Keys</div>
                </div>
                <div className="stat-item" onClick={() => navigateTo('/dids')}>
                  <div className="stat-value">{dids.length}</div>
                  <div className="stat-label">DIDs</div>
                </div>
                <div className="stat-item" onClick={() => navigateTo('/credentials')}>
                  <div className="stat-value">{credentials.length}</div>
                  <div className="stat-label">Credentials</div>
                </div>
                <div className="stat-item" onClick={() => navigateTo('/events')}>
                  <div className="stat-value">{events.length}</div>
                  <div className="stat-label">Events</div>
                </div>
              </div>
            </div>
          </div>

          <div className="status-card">
            <div className="card-header">
              <h3 className="card-title">Quick Actions</h3>
            </div>
            <div className="card-content">
              <div className="quick-actions-grid">
                {quickActions.map((action) => (
                  <button
                    key={action.path}
                    className="quick-action"
                    onClick={() => navigateTo(action.path)}
                    title={action.description}
                  >
                    <div className="action-icon">{action.icon}</div>
                    <div className="action-label">{action.label}</div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}