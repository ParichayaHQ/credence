import React, { useState } from 'react';

interface DIDCardProps {
  didData: any;
}

export function DIDCard({ didData }: DIDCardProps): JSX.Element {
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

  const getMethodIcon = (method: string) => {
    switch (method) {
      case 'key':
        return '🔐';
      default:
        return '🆔';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'deactivated':
        return 'warning';
      case 'revoked':
        return 'error';
      default:
        return 'neutral';
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      console.log(`${label} copied to clipboard`);
    });
  };

  const shortenDID = (did: string, length = 20) => {
    if (did.length <= length) return did;
    return `${did.substring(0, length)}...${did.substring(did.length - 10)}`;
  };

  return (
    <div className="did-card">
      <div className="did-card-header">
        <div className="did-info">
          <div className="did-icon">
            {getMethodIcon(didData.method)}
          </div>
          <div className="did-details">
            <div className="did-identifier" title={didData.did}>
              {shortenDID(didData.did)}
            </div>
            <div className="did-method">did:{didData.method}</div>
          </div>
        </div>
        <div className="did-actions">
          <div className={`status-badge ${getStatusColor(didData.status)}`}>
            {didData.status}
          </div>
          <button
            className="action-button"
            onClick={() => setShowDetails(!showDetails)}
            title="Toggle details"
          >
            {showDetails ? '▲' : '▼'}
          </button>
        </div>
      </div>

      <div className="did-card-body">
        <div className="did-meta">
          <div className="did-meta-item">
            <span className="meta-label">Created:</span>
            <span className="meta-value">{formatDate(didData.created)}</span>
          </div>
          <div className="did-meta-item">
            <span className="meta-label">Key ID:</span>
            <span className="meta-value">{didData.keyId}</span>
          </div>
          {didData.updated !== didData.created && (
            <div className="did-meta-item">
              <span className="meta-label">Updated:</span>
              <span className="meta-value">{formatDate(didData.updated)}</span>
            </div>
          )}
        </div>

        <div className="did-actions-bar">
          <button
            className="action-button-small"
            onClick={() => copyToClipboard(didData.did, 'DID')}
            title="Copy DID to clipboard"
          >
            📋 Copy DID
          </button>
        </div>

        {showDetails && (
          <div className="did-details-expanded">
            <div className="did-section">
              <div className="section-title">Full DID</div>
              <div className="did-full">
                <code className="did-code">{didData.did}</code>
              </div>
            </div>

            {didData.document && (
              <div className="did-section">
                <div className="section-title">DID Document</div>
                <div className="did-document">
                  <pre className="document-content">
                    {JSON.stringify(didData.document, null, 2)}
                  </pre>
                  <button
                    className="copy-button"
                    onClick={() => copyToClipboard(
                      JSON.stringify(didData.document, null, 2),
                      'DID Document'
                    )}
                  >
                    📋 Copy Document
                  </button>
                </div>
              </div>
            )}

            {didData.metadata && Object.keys(didData.metadata).length > 0 && (
              <div className="did-section">
                <div className="section-title">Metadata</div>
                <div className="metadata-list">
                  {Object.entries(didData.metadata).map(([key, value]) => (
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