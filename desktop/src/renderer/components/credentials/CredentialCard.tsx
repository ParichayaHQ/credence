import React, { useState } from 'react';

interface CredentialCardProps {
  credentialData: any;
  onDelete: () => void;
  isDeleting: boolean;
  isInBatchMode?: boolean;
  isSelected?: boolean;
  onSelect?: (selected: boolean) => void;
}

export function CredentialCard({ credentialData, onDelete, isDeleting, isInBatchMode = false, isSelected = false, onSelect }: CredentialCardProps): JSX.Element {
  const [showDetails, setShowDetails] = useState(false);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const getCredentialIcon = (types: string[]) => {
    if (!types || types.length === 0) return '📜';
    
    const type = types[0]?.toLowerCase() || '';
    if (type.includes('education') || type.includes('degree')) return '🎓';
    if (type.includes('employment') || type.includes('work')) return '💼';
    if (type.includes('certificate') || type.includes('certification')) return '🏆';
    if (type.includes('license')) return '📋';
    if (type.includes('identity') || type.includes('id')) return '🆔';
    return '📜';
  };

  const getStatusColor = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'valid':
        return 'success';
      case 'expired':
        return 'warning';
      case 'revoked':
      case 'suspended':
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

  const exportCredential = (credential: any) => {
    const dataStr = JSON.stringify(credential, null, 2);
    const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
    
    const exportFileDefaultName = `credential_${credential.id?.slice(0, 8) || 'export'}_${new Date().toISOString().split('T')[0]}.json`;
    
    const linkElement = document.createElement('a');
    linkElement.setAttribute('href', dataUri);
    linkElement.setAttribute('download', exportFileDefaultName);
    linkElement.click();
    linkElement.remove();
  };

  const getIssuerName = (issuer: any) => {
    if (typeof issuer === 'string') return issuer;
    if (typeof issuer === 'object' && issuer.name) return issuer.name;
    if (typeof issuer === 'object' && issuer.id) return issuer.id;
    return 'Unknown Issuer';
  };

  const getSubjectId = (subject: any) => {
    if (typeof subject === 'object' && subject.id) return subject.id;
    return 'Unknown Subject';
  };

  const isExpired = () => {
    if (!credentialData.expirationDate) return false;
    return new Date(credentialData.expirationDate) < new Date();
  };

  return (
    <div className={`credential-card ${isInBatchMode ? 'batch-mode' : ''} ${isSelected ? 'selected' : ''}`}>
      {isInBatchMode && (
        <div className="credential-checkbox">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={(e) => onSelect?.(e.target.checked)}
            className="batch-checkbox"
          />
        </div>
      )}
      <div className="credential-card-header">
        <div className="credential-info">
          <div className="credential-icon">
            {getCredentialIcon(credentialData.type)}
          </div>
          <div className="credential-details">
            <div className="credential-type">
              {credentialData.type?.[0] || 'Unknown Credential'}
            </div>
            <div className="credential-issuer">
              Issued by: {getIssuerName(credentialData.credential?.issuer)}
            </div>
          </div>
        </div>
        <div className="credential-actions">
          <div className={`status-badge ${getStatusColor(credentialData.status)}`}>
            {isExpired() ? 'expired' : (credentialData.status || 'valid')}
          </div>
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
            title="Delete credential"
          >
            {isDeleting ? '⏳' : '🗑️'}
          </button>
        </div>
      </div>

      <div className="credential-card-body">
        <div className="credential-meta">
          <div className="credential-meta-item">
            <span className="meta-label">Issued:</span>
            <span className="meta-value">{formatDate(credentialData.issuanceDate)}</span>
          </div>
          {credentialData.expirationDate && (
            <div className="credential-meta-item">
              <span className="meta-label">Expires:</span>
              <span className={`meta-value ${isExpired() ? 'expired' : ''}`}>
                {formatDate(credentialData.expirationDate)}
              </span>
            </div>
          )}
          <div className="credential-meta-item">
            <span className="meta-label">Subject:</span>
            <span className="meta-value subject-id">
              {getSubjectId(credentialData.credential?.credentialSubject)}
            </span>
          </div>
        </div>

        {credentialData.metadata?.tags && credentialData.metadata.tags.length > 0 && (
          <div className="credential-tags">
            {credentialData.metadata.tags.map((tag: string, index: number) => (
              <span key={index} className="tag">
                {tag}
              </span>
            ))}
          </div>
        )}

        <div className="credential-actions-bar">
          <button
            className="action-button-small"
            onClick={() => copyToClipboard(credentialData.id, 'Credential ID')}
            title="Copy credential ID"
          >
            📋 Copy ID
          </button>
          <button
            className="action-button-small"
            onClick={() => exportCredential(credentialData)}
            title="Export credential"
          >
            📤 Export
          </button>
        </div>

        {showDetails && (
          <div className="credential-details-expanded">
            <div className="credential-section">
              <div className="section-title">Credential Data</div>
              <div className="credential-json">
                <pre className="json-content">
                  {JSON.stringify(credentialData.credential, null, 2)}
                </pre>
                <button
                  className="copy-button"
                  onClick={() => copyToClipboard(
                    JSON.stringify(credentialData.credential, null, 2),
                    'Credential JSON'
                  )}
                >
                  📋 Copy JSON
                </button>
              </div>
            </div>

            {credentialData.metadata?.description && (
              <div className="credential-section">
                <div className="section-title">Description</div>
                <div className="description-text">
                  {credentialData.metadata.description}
                </div>
              </div>
            )}

            <div className="credential-section">
              <div className="section-title">Metadata</div>
              <div className="metadata-list">
                <div className="metadata-item">
                  <span className="metadata-key">ID:</span>
                  <span className="metadata-value">{credentialData.id}</span>
                </div>
                <div className="metadata-item">
                  <span className="metadata-key">Created:</span>
                  <span className="metadata-value">{formatDate(credentialData.created)}</span>
                </div>
                {credentialData.metadata?.importedAt && (
                  <div className="metadata-item">
                    <span className="metadata-key">Imported:</span>
                    <span className="metadata-value">{formatDate(credentialData.metadata.importedAt)}</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}