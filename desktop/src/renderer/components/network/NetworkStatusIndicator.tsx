import React, { useState } from 'react';
import { 
  useNetworkStatus, 
  useNetworkHealth, 
  useNetworkStats,
  useLatestCheckpoint 
} from '../../hooks/useWalletAPI';

interface NetworkStatusIndicatorProps {
  showDetails?: boolean;
}

interface NetworkStatusDetails {
  connected: boolean;
  peers: number;
  latency: number;
  lastCheckpoint?: string;
  health: 'healthy' | 'warning' | 'error';
}

export function NetworkStatusIndicator({ showDetails = false }: NetworkStatusIndicatorProps) {
  const [showTooltip, setShowTooltip] = useState(false);
  
  const { data: networkStatus, isLoading: statusLoading } = useNetworkStatus();
  const { data: networkHealth } = useNetworkHealth();
  const { data: networkStats } = useNetworkStats();
  const { data: latestCheckpoint } = useLatestCheckpoint();

  const getStatusDetails = (): NetworkStatusDetails => {
    const connected = networkStatus?.connected ?? false;
    const peers = networkStats?.peerCount ?? 0;
    const latency = networkHealth?.averageLatency ?? 0;
    
    let health: 'healthy' | 'warning' | 'error' = 'healthy';
    if (!connected || peers === 0) {
      health = 'error';
    } else if (peers < 3 || latency > 1000) {
      health = 'warning';
    }

    return {
      connected,
      peers,
      latency,
      lastCheckpoint: latestCheckpoint?.epoch,
      health
    };
  };

  const statusDetails = getStatusDetails();

  const getStatusIcon = () => {
    if (statusLoading) return '⟳';
    
    switch (statusDetails.health) {
      case 'healthy': return '●';
      case 'warning': return '◐';
      case 'error': return '○';
      default: return '○';
    }
  };

  const getStatusColor = () => {
    switch (statusDetails.health) {
      case 'healthy': return '#22c55e'; // green
      case 'warning': return '#f59e0b'; // amber
      case 'error': return '#ef4444'; // red
      default: return '#6b7280'; // gray
    }
  };

  const getStatusText = () => {
    if (statusLoading) return 'Connecting...';
    
    if (!statusDetails.connected) return 'Disconnected';
    if (statusDetails.peers === 0) return 'No Peers';
    if (statusDetails.health === 'warning') return 'Limited Connection';
    return 'Connected';
  };

  const formatLatency = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  };

  const renderTooltip = () => (
    <div className="network-status-tooltip">
      <div className="tooltip-row">
        <span className="tooltip-label">Status:</span>
        <span className="tooltip-value">{getStatusText()}</span>
      </div>
      <div className="tooltip-row">
        <span className="tooltip-label">Peers:</span>
        <span className="tooltip-value">{statusDetails.peers}</span>
      </div>
      <div className="tooltip-row">
        <span className="tooltip-label">Latency:</span>
        <span className="tooltip-value">{formatLatency(statusDetails.latency)}</span>
      </div>
      {statusDetails.lastCheckpoint && (
        <div className="tooltip-row">
          <span className="tooltip-label">Last Checkpoint:</span>
          <span className="tooltip-value">#{statusDetails.lastCheckpoint}</span>
        </div>
      )}
    </div>
  );

  return (
    <div 
      className="network-status-indicator"
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <div className="status-icon-container">
        <span 
          className={`status-icon ${statusLoading ? 'loading' : ''}`}
          style={{ color: getStatusColor() }}
        >
          {getStatusIcon()}
        </span>
        {showDetails && (
          <span className="status-text">
            {getStatusText()}
          </span>
        )}
      </div>
      
      {showTooltip && renderTooltip()}
    </div>
  );
}

export default NetworkStatusIndicator;