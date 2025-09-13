import React, { useState } from 'react';
import { 
  useNetworkPeers, 
  useNetworkStats 
} from '../../hooks/useWalletAPI';
import { useNotification } from '../../contexts/NotificationContext';

interface PeerConnection {
  id: string;
  address: string;
  multiaddr: string;
  status: 'connected' | 'connecting' | 'disconnected' | 'blocked';
  direction: 'inbound' | 'outbound';
  latency: number;
  lastSeen: string;
  connectedAt: string;
  protocols: string[];
  bytesIn: number;
  bytesOut: number;
  messagesIn: number;
  messagesOut: number;
  quality: 'excellent' | 'good' | 'poor';
  location?: {
    country: string;
    city: string;
  };
}

interface ConnectionManagerProps {
  maxPeers?: number;
  preferredPeers?: string[];
}

export function PeerConnectionManager({ 
  maxPeers = 50, 
  preferredPeers = [] 
}: ConnectionManagerProps) {
  const [selectedPeer, setSelectedPeer] = useState<string | null>(null);
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'latency' | 'quality' | 'connectedAt'>('quality');
  const [showAddPeerDialog, setShowAddPeerDialog] = useState(false);
  const [newPeerAddress, setNewPeerAddress] = useState('');
  
  const { data: peers = [], isLoading: peersLoading, refetch: refetchPeers } = useNetworkPeers();
  const { data: networkStats } = useNetworkStats();
  const { showNotification } = useNotification();

  // Mock functions that would integrate with the P2P network
  const connectToPeer = async (address: string) => {
    try {
      // This would call the actual P2P connection API
      console.log('Connecting to peer:', address);
      showNotification(`Attempting to connect to ${address}`, 'info');
      // Simulate connection attempt
      setTimeout(() => {
        refetchPeers();
        showNotification('Peer connection initiated', 'success');
      }, 1000);
    } catch (error) {
      showNotification('Failed to connect to peer', 'error');
    }
  };

  const disconnectFromPeer = async (peerId: string) => {
    try {
      console.log('Disconnecting from peer:', peerId);
      showNotification('Disconnecting from peer...', 'info');
      // Simulate disconnection
      setTimeout(() => {
        refetchPeers();
        showNotification('Peer disconnected', 'success');
      }, 500);
    } catch (error) {
      showNotification('Failed to disconnect from peer', 'error');
    }
  };

  const blockPeer = async (peerId: string) => {
    try {
      console.log('Blocking peer:', peerId);
      showNotification('Blocking peer...', 'info');
      setTimeout(() => {
        refetchPeers();
        showNotification('Peer blocked', 'success');
      }, 500);
    } catch (error) {
      showNotification('Failed to block peer', 'error');
    }
  };

  const unblockPeer = async (peerId: string) => {
    try {
      console.log('Unblocking peer:', peerId);
      showNotification('Unblocking peer...', 'info');
      setTimeout(() => {
        refetchPeers();
        showNotification('Peer unblocked', 'success');
      }, 500);
    } catch (error) {
      showNotification('Failed to unblock peer', 'error');
    }
  };

  const getConnectionQualityColor = (quality: string): string => {
    switch (quality) {
      case 'excellent': return '#22c55e';
      case 'good': return '#f59e0b';
      case 'poor': return '#ef4444';
      default: return '#6b7280';
    }
  };

  const getConnectionStatusIcon = (status: string): string => {
    switch (status) {
      case 'connected': return '🟢';
      case 'connecting': return '🟡';
      case 'disconnected': return '🔴';
      case 'blocked': return '🚫';
      default: return '⚪';
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / 1048576).toFixed(1)} MB`;
  };

  const formatDuration = (startTime: string): string => {
    const start = new Date(startTime);
    const now = new Date();
    const diffMs = now.getTime() - start.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffMinutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
    
    if (diffHours > 0) return `${diffHours}h ${diffMinutes}m`;
    return `${diffMinutes}m`;
  };

  const filteredPeers = peers
    .filter((peer: PeerConnection) => {
      if (filterStatus === 'all') return true;
      return peer.status === filterStatus;
    })
    .sort((a: PeerConnection, b: PeerConnection) => {
      switch (sortBy) {
        case 'latency':
          return a.latency - b.latency;
        case 'quality':
          const qualityOrder = { excellent: 0, good: 1, poor: 2 };
          return qualityOrder[a.quality] - qualityOrder[b.quality];
        case 'connectedAt':
          return new Date(b.connectedAt).getTime() - new Date(a.connectedAt).getTime();
        default:
          return 0;
      }
    });

  const renderConnectionSummary = () => (
    <div className="connection-summary">
      <div className="summary-stats">
        <div className="stat-item">
          <span className="stat-value">{peers.filter(p => p.status === 'connected').length}</span>
          <span className="stat-label">Connected</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{peers.filter(p => p.status === 'connecting').length}</span>
          <span className="stat-label">Connecting</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{peers.filter(p => p.status === 'blocked').length}</span>
          <span className="stat-label">Blocked</span>
        </div>
        <div className="stat-item">
          <span className="stat-value">{maxPeers}</span>
          <span className="stat-label">Max Peers</span>
        </div>
      </div>
      
      <div className="network-metrics">
        <div className="metric-item">
          <span className="metric-label">Total Bandwidth:</span>
          <span className="metric-value">
            ↑ {formatBytes(networkStats?.totalBytesOut || 0)} / 
            ↓ {formatBytes(networkStats?.totalBytesIn || 0)}
          </span>
        </div>
        <div className="metric-item">
          <span className="metric-label">Messages/sec:</span>
          <span className="metric-value">{networkStats?.messagesPerSecond || 0}</span>
        </div>
      </div>
    </div>
  );

  const renderAddPeerDialog = () => {
    if (!showAddPeerDialog) return null;

    return (
      <div className="add-peer-dialog">
        <div className="dialog-content">
          <div className="dialog-header">
            <h4>Add Peer Connection</h4>
            <button 
              onClick={() => setShowAddPeerDialog(false)}
              className="close-button"
            >
              ×
            </button>
          </div>
          
          <div className="dialog-body">
            <div className="form-group">
              <label htmlFor="peer-address">Peer Multiaddress:</label>
              <input
                id="peer-address"
                type="text"
                value={newPeerAddress}
                onChange={(e) => setNewPeerAddress(e.target.value)}
                placeholder="/ip4/192.168.1.100/tcp/4001/p2p/12D3KooW..."
                className="peer-address-input"
              />
              <div className="input-hint">
                Enter a valid libp2p multiaddress for the peer
              </div>
            </div>
          </div>
          
          <div className="dialog-actions">
            <button
              onClick={() => {
                if (newPeerAddress.trim()) {
                  connectToPeer(newPeerAddress.trim());
                  setNewPeerAddress('');
                  setShowAddPeerDialog(false);
                }
              }}
              disabled={!newPeerAddress.trim()}
              className="connect-button primary"
            >
              Connect
            </button>
            <button
              onClick={() => setShowAddPeerDialog(false)}
              className="cancel-button"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    );
  };

  const renderPeerDetails = (peer: PeerConnection) => (
    <div className="peer-details-panel">
      <div className="details-header">
        <h4>Peer Details</h4>
        <button 
          onClick={() => setSelectedPeer(null)}
          className="close-details"
        >
          ×
        </button>
      </div>
      
      <div className="details-content">
        <div className="detail-section">
          <h5>Connection Info</h5>
          <div className="detail-item">
            <span className="label">ID:</span>
            <span className="value monospace">{peer.id}</span>
          </div>
          <div className="detail-item">
            <span className="label">Address:</span>
            <span className="value">{peer.address}</span>
          </div>
          <div className="detail-item">
            <span className="label">Multiaddr:</span>
            <span className="value monospace">{peer.multiaddr}</span>
          </div>
          <div className="detail-item">
            <span className="label">Direction:</span>
            <span className="value">{peer.direction}</span>
          </div>
          <div className="detail-item">
            <span className="label">Connected:</span>
            <span className="value">{formatDuration(peer.connectedAt)}</span>
          </div>
        </div>

        <div className="detail-section">
          <h5>Performance</h5>
          <div className="detail-item">
            <span className="label">Latency:</span>
            <span className="value">{peer.latency}ms</span>
          </div>
          <div className="detail-item">
            <span className="label">Quality:</span>
            <span 
              className="value quality-indicator"
              style={{ color: getConnectionQualityColor(peer.quality) }}
            >
              {peer.quality.toUpperCase()}
            </span>
          </div>
        </div>

        <div className="detail-section">
          <h5>Traffic Statistics</h5>
          <div className="detail-item">
            <span className="label">Data In:</span>
            <span className="value">{formatBytes(peer.bytesIn)}</span>
          </div>
          <div className="detail-item">
            <span className="label">Data Out:</span>
            <span className="value">{formatBytes(peer.bytesOut)}</span>
          </div>
          <div className="detail-item">
            <span className="label">Messages In:</span>
            <span className="value">{peer.messagesIn.toLocaleString()}</span>
          </div>
          <div className="detail-item">
            <span className="label">Messages Out:</span>
            <span className="value">{peer.messagesOut.toLocaleString()}</span>
          </div>
        </div>

        <div className="detail-section">
          <h5>Protocols</h5>
          <div className="protocols-list">
            {peer.protocols.map((protocol, index) => (
              <span key={index} className="protocol-tag">
                {protocol}
              </span>
            ))}
          </div>
        </div>

        {peer.location && (
          <div className="detail-section">
            <h5>Location</h5>
            <div className="detail-item">
              <span className="label">Country:</span>
              <span className="value">{peer.location.country}</span>
            </div>
            <div className="detail-item">
              <span className="label">City:</span>
              <span className="value">{peer.location.city}</span>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  const renderPeersList = () => (
    <div className="peers-list">
      <div className="list-header">
        <div className="list-controls">
          <div className="filter-controls">
            <label>Status:</label>
            <select 
              value={filterStatus} 
              onChange={(e) => setFilterStatus(e.target.value)}
              className="filter-select"
            >
              <option value="all">All</option>
              <option value="connected">Connected</option>
              <option value="connecting">Connecting</option>
              <option value="disconnected">Disconnected</option>
              <option value="blocked">Blocked</option>
            </select>
          </div>
          
          <div className="sort-controls">
            <label>Sort by:</label>
            <select 
              value={sortBy} 
              onChange={(e) => setSortBy(e.target.value as any)}
              className="sort-select"
            >
              <option value="quality">Quality</option>
              <option value="latency">Latency</option>
              <option value="connectedAt">Connection Time</option>
            </select>
          </div>
          
          <button 
            onClick={() => refetchPeers()}
            disabled={peersLoading}
            className="refresh-button"
          >
            {peersLoading ? '⟳' : '↻'} Refresh
          </button>
        </div>
      </div>
      
      <div className="peers-table">
        <div className="table-header">
          <div className="col-status">Status</div>
          <div className="col-peer">Peer</div>
          <div className="col-latency">Latency</div>
          <div className="col-quality">Quality</div>
          <div className="col-traffic">Traffic</div>
          <div className="col-actions">Actions</div>
        </div>
        
        <div className="table-body">
          {filteredPeers.map((peer: PeerConnection) => (
            <div key={peer.id} className="table-row">
              <div className="col-status">
                <span className="status-indicator">
                  {getConnectionStatusIcon(peer.status)}
                  {peer.status}
                </span>
              </div>
              
              <div className="col-peer">
                <div className="peer-info">
                  <div className="peer-id">{peer.id.substring(0, 16)}...</div>
                  <div className="peer-address">{peer.address}</div>
                </div>
              </div>
              
              <div className="col-latency">
                {peer.latency}ms
              </div>
              
              <div className="col-quality">
                <span 
                  className="quality-badge"
                  style={{ color: getConnectionQualityColor(peer.quality) }}
                >
                  {peer.quality}
                </span>
              </div>
              
              <div className="col-traffic">
                <div className="traffic-info">
                  <div>↓ {formatBytes(peer.bytesIn)}</div>
                  <div>↑ {formatBytes(peer.bytesOut)}</div>
                </div>
              </div>
              
              <div className="col-actions">
                <button
                  onClick={() => setSelectedPeer(peer.id)}
                  className="action-button details"
                  title="View details"
                >
                  👁
                </button>
                
                {peer.status === 'connected' ? (
                  <button
                    onClick={() => disconnectFromPeer(peer.id)}
                    className="action-button disconnect"
                    title="Disconnect"
                  >
                    🔌
                  </button>
                ) : peer.status === 'disconnected' ? (
                  <button
                    onClick={() => connectToPeer(peer.multiaddr)}
                    className="action-button connect"
                    title="Connect"
                  >
                    🔗
                  </button>
                ) : null}
                
                {peer.status !== 'blocked' ? (
                  <button
                    onClick={() => blockPeer(peer.id)}
                    className="action-button block"
                    title="Block peer"
                  >
                    🚫
                  </button>
                ) : (
                  <button
                    onClick={() => unblockPeer(peer.id)}
                    className="action-button unblock"
                    title="Unblock peer"
                  >
                    ✅
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );

  if (peersLoading) {
    return (
      <div className="peer-connection-manager loading">
        <div className="loading-spinner">Loading peer connections...</div>
      </div>
    );
  }

  return (
    <div className="peer-connection-manager">
      <div className="manager-header">
        <h3>Peer Connection Manager</h3>
        <button
          onClick={() => setShowAddPeerDialog(true)}
          className="add-peer-button"
        >
          + Add Peer
        </button>
      </div>

      {renderConnectionSummary()}
      {renderPeersList()}
      
      {selectedPeer && (
        <div className="peer-details-overlay">
          {renderPeerDetails(peers.find((p: PeerConnection) => p.id === selectedPeer)!)}
        </div>
      )}
      
      {renderAddPeerDialog()}
    </div>
  );
}

export default PeerConnectionManager;