import React, { useState } from 'react';
import NetworkHealthMonitor from '../components/network/NetworkHealthMonitor';
import CheckpointVerificationDisplay from '../components/network/CheckpointVerificationDisplay';
import RealTimeEventFeed from '../components/network/RealTimeEventFeed';
import PeerConnectionManager from '../components/network/PeerConnectionManager';
import RulesRegistryManager from '../components/network/RulesRegistryManager';

type NetworkTab = 'overview' | 'events' | 'checkpoints' | 'peers' | 'rules';

export function Network() {
  const [activeTab, setActiveTab] = useState<NetworkTab>('overview');

  const tabs = [
    { id: 'overview' as const, label: 'Network Overview', icon: '📊' },
    { id: 'events' as const, label: 'Live Events', icon: '📡' },
    { id: 'checkpoints' as const, label: 'Checkpoints', icon: '📋' },
    { id: 'peers' as const, label: 'Peer Management', icon: '🌐' },
    { id: 'rules' as const, label: 'Rules Registry', icon: '⚙️' },
  ];

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return <NetworkHealthMonitor />;
      case 'events':
        return (
          <div className="events-tab">
            <RealTimeEventFeed 
              maxEvents={100}
              autoScroll={true}
              showFilters={true}
            />
          </div>
        );
      case 'checkpoints':
        return <CheckpointVerificationDisplay />;
      case 'peers':
        return (
          <PeerConnectionManager 
            maxPeers={50}
            preferredPeers={[]}
          />
        );
      case 'rules':
        return <RulesRegistryManager />;
      default:
        return <NetworkHealthMonitor />;
    }
  };

  return (
    <div className="network-page">
      <div className="page-header">
        <h2>Network Management</h2>
        <p className="page-description">
          Monitor network health, manage peer connections, verify checkpoints, and view live events.
        </p>
      </div>

      <div className="network-tabs">
        <div className="tab-navigation">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`tab-button ${activeTab === tab.id ? 'active' : ''}`}
            >
              <span className="tab-icon">{tab.icon}</span>
              <span className="tab-label">{tab.label}</span>
            </button>
          ))}
        </div>

        <div className="tab-content">
          {renderTabContent()}
        </div>
      </div>
    </div>
  );
}

export default Network;