import React, { useState } from 'react';
import { useWallet } from '../../contexts/WalletContext';
import NetworkStatusIndicator from '../network/NetworkStatusIndicator';
import GlobalSearch, { SearchTrigger } from '../search/GlobalSearch';

interface HeaderProps {
  isLocked: boolean;
  networkConnected: boolean;
}

export function Header({ isLocked, networkConnected }: HeaderProps): JSX.Element {
  const { getNetworkStatus } = useWallet();
  const [searchOpen, setSearchOpen] = useState(false);

  const handleRefreshNetwork = () => {
    getNetworkStatus();
  };

  return (
    <header className="header">
      <div className="header-left">
        <h1 className="app-title">Credence Wallet</h1>
      </div>

      <div className="header-center">
        {/* Global search */}
        <div className="search-container">
          <SearchTrigger onOpen={() => setSearchOpen(true)} />
        </div>
        
        {/* Enhanced network status indicators */}
        <div className="status-indicators">
          <NetworkStatusIndicator showDetails={true} />
        </div>
      </div>

      <div className="header-right">
        <div className="wallet-status">
          <div className={`lock-indicator ${isLocked ? 'locked' : 'unlocked'}`}>
            <span className="lock-icon">
              {isLocked ? '🔒' : '🔓'}
            </span>
            <span className="lock-text">
              {isLocked ? 'Locked' : 'Unlocked'}
            </span>
          </div>
        </div>

        <div className="window-controls">
          <button 
            className="window-control minimize"
            onClick={() => window.walletAPI?.window.minimize()}
            title="Minimize"
          >
            −
          </button>
          <button 
            className="window-control close"
            onClick={() => window.walletAPI?.window.close()}
            title="Close"
          >
            ×
          </button>
        </div>
      </div>

      {/* Global Search Modal */}
      <GlobalSearch 
        isOpen={searchOpen} 
        onClose={() => setSearchOpen(false)} 
      />
    </header>
  );
}