import React, { useState } from 'react';
import { 
  useCheckpoints, 
  useLatestCheckpoint, 
  useVerifyCheckpoint 
} from '../../hooks/useWalletAPI';

interface Checkpoint {
  epoch: string;
  root: string;
  timestamp: string;
  signers: string[];
  signature: string;
  treeSize: number;
  verified: boolean;
  verificationTime?: string;
}

export function CheckpointVerificationDisplay() {
  const [selectedCheckpoint, setSelectedCheckpoint] = useState<string | null>(null);
  const [verificationDetails, setVerificationDetails] = useState<any>(null);
  
  const { data: checkpoints = [], isLoading: checkpointsLoading, refetch: refetchCheckpoints } = useCheckpoints();
  const { data: latestCheckpoint, isLoading: latestLoading } = useLatestCheckpoint();
  const verifyCheckpointMutation = useVerifyCheckpoint();

  const handleVerifyCheckpoint = async (epoch: string) => {
    try {
      const result = await verifyCheckpointMutation.mutateAsync(epoch);
      setVerificationDetails(result);
      setSelectedCheckpoint(epoch);
      // Refresh checkpoints to update verification status
      refetchCheckpoints();
    } catch (error) {
      console.error('Verification failed:', error);
    }
  };

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  const formatRoot = (root: string): string => {
    return `${root.substring(0, 8)}...${root.substring(root.length - 8)}`;
  };

  const getVerificationStatusIcon = (checkpoint: Checkpoint) => {
    if (checkpoint.verified) return '✅';
    if (verifyCheckpointMutation.isPending && selectedCheckpoint === checkpoint.epoch) return '⟳';
    return '⏸';
  };

  const getVerificationStatusColor = (checkpoint: Checkpoint) => {
    if (checkpoint.verified) return '#22c55e';
    if (verifyCheckpointMutation.isPending && selectedCheckpoint === checkpoint.epoch) return '#f59e0b';
    return '#6b7280';
  };

  const renderLatestCheckpoint = () => {
    if (latestLoading) {
      return <div className="latest-checkpoint loading">Loading latest checkpoint...</div>;
    }

    if (!latestCheckpoint) {
      return <div className="latest-checkpoint error">No checkpoint data available</div>;
    }

    return (
      <div className="latest-checkpoint">
        <div className="checkpoint-header">
          <h4>Latest Checkpoint</h4>
          <span className="checkpoint-epoch">#{latestCheckpoint.epoch}</span>
        </div>
        
        <div className="checkpoint-details">
          <div className="detail-row">
            <span className="label">Root Hash:</span>
            <span className="value monospace">{formatRoot(latestCheckpoint.root)}</span>
          </div>
          <div className="detail-row">
            <span className="label">Tree Size:</span>
            <span className="value">{latestCheckpoint.treeSize?.toLocaleString()}</span>
          </div>
          <div className="detail-row">
            <span className="label">Timestamp:</span>
            <span className="value">{formatTimestamp(latestCheckpoint.timestamp)}</span>
          </div>
          <div className="detail-row">
            <span className="label">Signers:</span>
            <span className="value">{latestCheckpoint.signers?.length || 0} nodes</span>
          </div>
        </div>

        <div className="checkpoint-actions">
          <button
            onClick={() => handleVerifyCheckpoint(latestCheckpoint.epoch)}
            disabled={verifyCheckpointMutation.isPending}
            className="verify-button primary"
          >
            {verifyCheckpointMutation.isPending && selectedCheckpoint === latestCheckpoint.epoch 
              ? 'Verifying...' 
              : 'Verify Checkpoint'
            }
          </button>
        </div>
      </div>
    );
  };

  const renderCheckpointsList = () => {
    if (checkpointsLoading) {
      return <div className="checkpoints-list loading">Loading checkpoints...</div>;
    }

    if (checkpoints.length === 0) {
      return <div className="checkpoints-list empty">No checkpoints available</div>;
    }

    return (
      <div className="checkpoints-list">
        <div className="list-header">
          <h4>Recent Checkpoints</h4>
          <button onClick={() => refetchCheckpoints()} className="refresh-button">
            ↻ Refresh
          </button>
        </div>
        
        <div className="checkpoints-table">
          <div className="table-header">
            <div className="column-epoch">Epoch</div>
            <div className="column-timestamp">Timestamp</div>
            <div className="column-root">Root Hash</div>
            <div className="column-size">Tree Size</div>
            <div className="column-status">Status</div>
            <div className="column-actions">Actions</div>
          </div>
          
          <div className="table-body">
            {checkpoints.map((checkpoint: Checkpoint) => (
              <div key={checkpoint.epoch} className="table-row">
                <div className="column-epoch">#{checkpoint.epoch}</div>
                <div className="column-timestamp">
                  {formatTimestamp(checkpoint.timestamp)}
                </div>
                <div className="column-root monospace">
                  {formatRoot(checkpoint.root)}
                </div>
                <div className="column-size">
                  {checkpoint.treeSize?.toLocaleString()}
                </div>
                <div className="column-status">
                  <span 
                    className="status-indicator"
                    style={{ color: getVerificationStatusColor(checkpoint) }}
                  >
                    {getVerificationStatusIcon(checkpoint)}
                    {checkpoint.verified ? 'Verified' : 'Pending'}
                  </span>
                </div>
                <div className="column-actions">
                  <button
                    onClick={() => handleVerifyCheckpoint(checkpoint.epoch)}
                    disabled={verifyCheckpointMutation.isPending || checkpoint.verified}
                    className="verify-button small"
                  >
                    {verifyCheckpointMutation.isPending && selectedCheckpoint === checkpoint.epoch
                      ? '⟳'
                      : checkpoint.verified 
                        ? '✓' 
                        : 'Verify'
                    }
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  };

  const renderVerificationDetails = () => {
    if (!verificationDetails || !selectedCheckpoint) return null;

    return (
      <div className="verification-details">
        <div className="details-header">
          <h4>Verification Results - Checkpoint #{selectedCheckpoint}</h4>
          <button 
            onClick={() => setVerificationDetails(null)}
            className="close-button"
          >
            ×
          </button>
        </div>
        
        <div className="verification-result">
          <div className={`result-status ${verificationDetails.verified ? 'success' : 'error'}`}>
            <span className="status-icon">
              {verificationDetails.verified ? '✅' : '❌'}
            </span>
            <span className="status-text">
              {verificationDetails.verified ? 'Verification Successful' : 'Verification Failed'}
            </span>
          </div>

          {verificationDetails.details && (
            <div className="verification-details-content">
              <div className="detail-section">
                <h5>Signature Verification</h5>
                <div className="detail-item">
                  <span className="label">Valid Signature:</span>
                  <span className={`value ${verificationDetails.details.signatureValid ? 'success' : 'error'}`}>
                    {verificationDetails.details.signatureValid ? 'Yes' : 'No'}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="label">Threshold Met:</span>
                  <span className={`value ${verificationDetails.details.thresholdMet ? 'success' : 'error'}`}>
                    {verificationDetails.details.thresholdMet ? 'Yes' : 'No'}
                  </span>
                </div>
              </div>

              <div className="detail-section">
                <h5>Consensus Information</h5>
                <div className="detail-item">
                  <span className="label">Participating Signers:</span>
                  <span className="value">{verificationDetails.details.participatingSigners}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Required Threshold:</span>
                  <span className="value">{verificationDetails.details.requiredThreshold}</span>
                </div>
              </div>

              {verificationDetails.errors && verificationDetails.errors.length > 0 && (
                <div className="detail-section error">
                  <h5>Verification Errors</h5>
                  {verificationDetails.errors.map((error: string, index: number) => (
                    <div key={index} className="error-item">
                      {error}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          <div className="verification-timestamp">
            Verified at: {formatTimestamp(verificationDetails.timestamp || new Date().toISOString())}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="checkpoint-verification-display">
      <div className="display-header">
        <h3>Checkpoint Verification</h3>
        <p className="description">
          Verify transparency log checkpoints to ensure data integrity and consensus.
        </p>
      </div>

      <div className="display-content">
        {renderLatestCheckpoint()}
        {renderCheckpointsList()}
        {renderVerificationDetails()}
      </div>
    </div>
  );
}

export default CheckpointVerificationDisplay;