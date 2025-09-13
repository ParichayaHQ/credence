import React, { useState } from 'react';
import { 
  useNetworkHealth, 
  useNetworkPeers, 
  useNetworkStats 
} from '../../hooks/useWalletAPI';

interface NetworkPeer {
  id: string;
  address: string;
  latency: number;
  status: 'connected' | 'connecting' | 'disconnected';
  lastSeen: string;
  protocols: string[];
}

interface NetworkHealthData {
  overall: 'healthy' | 'warning' | 'critical';
  connectivity: number; // 0-100 percentage
  performance: number; // 0-100 percentage
  reliability: number; // 0-100 percentage
  averageLatency: number;
  packetLoss: number;
  uptime: number;
}

interface NetworkStatsData {
  peerCount: number;
  messagesPerSecond: number;
  bytesPerSecond: number;
  totalMessages: number;
  totalBytes: number;
  uptime: number;
}

export function NetworkHealthMonitor() {
  const [selectedTab, setSelectedTab] = useState<'overview' | 'peers' | 'stats'>('overview');
  
  const { data: healthData, isLoading: healthLoading } = useNetworkHealth();
  const { data: peersData = [], isLoading: peersLoading } = useNetworkPeers();
  const { data: statsData, isLoading: statsLoading } = useNetworkStats();

  const formatBytes = (bytes: number): string => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  const formatUptime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const getHealthColor = (percentage: number): string => {
    if (percentage >= 80) return '#22c55e'; // green
    if (percentage >= 60) return '#f59e0b'; // amber
    return '#ef4444'; // red
  };

  const renderOverviewTab = () => (
    <div className="network-overview">
      <div className="health-metrics">
        <div className="metric-card">
          <div className="metric-header">
            <h4>Overall Health</h4>
            <span className={`health-badge ${healthData?.overall || 'warning'}`}>
              {healthData?.overall?.toUpperCase() || 'UNKNOWN'}
            </span>
          </div>
          <div className="metric-bars">
            <div className="metric-bar">
              <label>Connectivity</label>
              <div className="progress-bar">
                <div 
                  className="progress-fill" 
                  style={{ 
                    width: `${healthData?.connectivity || 0}%`,
                    backgroundColor: getHealthColor(healthData?.connectivity || 0)
                  }}
                />
              </div>
              <span>{healthData?.connectivity || 0}%</span>
            </div>
            
            <div className="metric-bar">
              <label>Performance</label>
              <div className="progress-bar">
                <div 
                  className="progress-fill" 
                  style={{ 
                    width: `${healthData?.performance || 0}%`,
                    backgroundColor: getHealthColor(healthData?.performance || 0)
                  }}
                />
              </div>
              <span>{healthData?.performance || 0}%</span>
            </div>
            
            <div className="metric-bar">
              <label>Reliability</label>
              <div className="progress-bar">
                <div 
                  className="progress-fill" 
                  style={{ 
                    width: `${healthData?.reliability || 0}%`,
                    backgroundColor: getHealthColor(healthData?.reliability || 0)
                  }}
                />
              </div>
              <span>{healthData?.reliability || 0}%</span>
            </div>
          </div>
        </div>

        <div className="stats-grid">
          <div className="stat-item">
            <span className="stat-label">Average Latency</span>
            <span className="stat-value">
              {healthData?.averageLatency || 0}ms
            </span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Packet Loss</span>
            <span className="stat-value">
              {(healthData?.packetLoss || 0).toFixed(1)}%
            </span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Uptime</span>
            <span className="stat-value">
              {formatUptime(healthData?.uptime || 0)}
            </span>
          </div>
        </div>
      </div>
    </div>
  );

  const renderPeersTab = () => (
    <div className="network-peers">
      <div className="peers-header">
        <h4>Connected Peers ({peersData.length})</h4>
        <button className="refresh-button" disabled={peersLoading}>
          {peersLoading ? '⟳' : '↻'} Refresh
        </button>
      </div>
      
      <div className="peers-list">
        {peersData.length === 0 ? (
          <div className="no-peers">No peers connected</div>
        ) : (
          peersData.map((peer: NetworkPeer) => (
            <div key={peer.id} className="peer-card">
              <div className="peer-header">
                <div className="peer-id">{peer.id.substring(0, 16)}...</div>
                <div className={`peer-status ${peer.status}`}>
                  {peer.status.toUpperCase()}
                </div>
              </div>
              
              <div className="peer-details">
                <div className="peer-detail">
                  <span className="label">Address:</span>
                  <span className="value">{peer.address}</span>
                </div>
                <div className="peer-detail">
                  <span className="label">Latency:</span>
                  <span className="value">{peer.latency}ms</span>
                </div>
                <div className="peer-detail">
                  <span className="label">Last Seen:</span>
                  <span className="value">{new Date(peer.lastSeen).toLocaleTimeString()}</span>
                </div>
                {peer.protocols.length > 0 && (
                  <div className="peer-protocols">
                    <span className="label">Protocols:</span>
                    <div className="protocols-list">
                      {peer.protocols.map((protocol) => (
                        <span key={protocol} className="protocol-tag">
                          {protocol}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );

  const renderStatsTab = () => (
    <div className="network-stats">
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-header">
            <h5>Network Activity</h5>
          </div>
          <div className="stat-content">
            <div className="stat-row">
              <span className="label">Messages/sec:</span>
              <span className="value">{statsData?.messagesPerSecond || 0}</span>
            </div>
            <div className="stat-row">
              <span className="label">Bandwidth:</span>
              <span className="value">{formatBytes(statsData?.bytesPerSecond || 0)}/s</span>
            </div>
            <div className="stat-row">
              <span className="label">Total Messages:</span>
              <span className="value">{(statsData?.totalMessages || 0).toLocaleString()}</span>
            </div>
            <div className="stat-row">
              <span className="label">Total Data:</span>
              <span className="value">{formatBytes(statsData?.totalBytes || 0)}</span>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-header">
            <h5>Connection Info</h5>
          </div>
          <div className="stat-content">
            <div className="stat-row">
              <span className="label">Peers Connected:</span>
              <span className="value">{statsData?.peerCount || 0}</span>
            </div>
            <div className="stat-row">
              <span className="label">Session Uptime:</span>
              <span className="value">{formatUptime(statsData?.uptime || 0)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  if (healthLoading && peersLoading && statsLoading) {
    return (
      <div className="network-health-monitor loading">
        <div className="loading-spinner">Loading network data...</div>
      </div>
    );
  }

  return (
    <div className="network-health-monitor">
      <div className="monitor-header">
        <h3>Network Health Monitor</h3>
        <div className="tab-navigation">
          <button 
            className={selectedTab === 'overview' ? 'active' : ''}
            onClick={() => setSelectedTab('overview')}
          >
            Overview
          </button>
          <button 
            className={selectedTab === 'peers' ? 'active' : ''}
            onClick={() => setSelectedTab('peers')}
          >
            Peers ({peersData.length})
          </button>
          <button 
            className={selectedTab === 'stats' ? 'active' : ''}
            onClick={() => setSelectedTab('stats')}
          >
            Statistics
          </button>
        </div>
      </div>

      <div className="monitor-content">
        {selectedTab === 'overview' && renderOverviewTab()}
        {selectedTab === 'peers' && renderPeersTab()}
        {selectedTab === 'stats' && renderStatsTab()}
      </div>
    </div>
  );
}

export default NetworkHealthMonitor;