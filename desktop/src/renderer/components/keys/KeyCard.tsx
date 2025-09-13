import React, { useState } from 'react';

interface KeyCardProps {
  keyData: any;
  onDelete: () => void;
  isDeleting: boolean;
}

export function KeyCard({ keyData, onDelete, isDeleting }: KeyCardProps): JSX.Element {
  const [showDetails, setShowDetails] = useState(false);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getKeyTypeIcon = (keyType: string) => {
    switch (keyType) {
      case 'Ed25519':
        return '🔐';
      case 'Secp256k1':
        return '₿';
      default:
        return '🔑';
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      // Could show a temporary tooltip here
      console.log(`${label} copied to clipboard`);
    });
  };

  return (
    <div className="key-card">
      <div className="key-card-header">
        <div className="key-info">
          <div className="key-icon">
            {getKeyTypeIcon(keyData.keyType)}
          </div>
          <div className="key-details">
            <div className="key-id">{keyData.id}</div>
            <div className="key-type">{keyData.keyType} • {keyData.algorithm}</div>
          </div>
        </div>
        <div className="key-actions">
          <button
            className="action-button"
            onClick={() => setShowDetails(!showDetails)}
            title="Toggle details"
          >
            {showDetails ? '▲' : '▼'}
          </button>
          <button
            className="action-button delete-button"
            onClick={onDelete}
            disabled={isDeleting}
            title="Delete key"
          >
            {isDeleting ? '⏳' : '🗑️'}
          </button>
        </div>
      </div>

      <div className="key-card-body">
        <div className="key-meta">
          <div className="key-meta-item">
            <span className="meta-label">Created:</span>
            <span className="meta-value">{formatDate(keyData.created)}</span>
          </div>
          <div className="key-meta-item">
            <span className="meta-label">Usage:</span>
            <span className="meta-value">
              {keyData.usage ? keyData.usage.join(', ') : 'General'}
            </span>
          </div>
        </div>

        {showDetails && (
          <div className="key-details-expanded">
            <div className="key-section">
              <div className="section-title">Public Key (JWK)</div>
              <div className="key-data">
                {keyData.publicKeyJwk ? (
                  <div className="jwk-display">
                    <pre className="jwk-content">
                      {JSON.stringify(keyData.publicKeyJwk, null, 2)}
                    </pre>
                    <button
                      className="copy-button"
                      onClick={() => copyToClipboard(
                        JSON.stringify(keyData.publicKeyJwk, null, 2),
                        'Public Key JWK'
                      )}
                    >
                      📋 Copy
                    </button>
                  </div>
                ) : (
                  <span className="no-data">Public key not available</span>
                )}
              </div>
            </div>

            {keyData.metadata && Object.keys(keyData.metadata).length > 0 && (
              <div className="key-section">
                <div className="section-title">Metadata</div>
                <div className="metadata-list">
                  {Object.entries(keyData.metadata).map(([key, value]) => (
                    <div key={key} className="metadata-item">
                      <span className="metadata-key">{key}:</span>
                      <span className="metadata-value">
                        {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}