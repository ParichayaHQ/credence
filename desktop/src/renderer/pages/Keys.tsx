import React, { useState } from 'react';
import { useKeys, useGenerateKey, useDeleteKey } from '../hooks/useWalletAPI';
import { KeyGenerationModal } from '../components/keys/KeyGenerationModal';
import { KeyCard } from '../components/keys/KeyCard';

export function Keys(): JSX.Element {
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const { data: keys = [], isLoading, error } = useKeys();
  const generateKey = useGenerateKey();
  const deleteKey = useDeleteKey();

  const handleGenerateKey = async (keyType: string) => {
    await generateKey.mutateAsync(keyType);
    setShowGenerateModal(false);
  };

  const handleDeleteKey = async (keyId: string, keyName: string) => {
    if (confirm(`Are you sure you want to delete the key "${keyName}"? This action cannot be undone.`)) {
      await deleteKey.mutateAsync(keyId);
    }
  };

  if (isLoading) {
    return (
      <div className="page keys">
        <div className="page-header">
          <h1 className="page-title">Cryptographic Keys</h1>
          <p className="page-subtitle">Manage your cryptographic keys for signing and encryption</p>
        </div>
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading keys...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page keys">
        <div className="page-header">
          <h1 className="page-title">Cryptographic Keys</h1>
          <p className="page-subtitle">Manage your cryptographic keys for signing and encryption</p>
        </div>
        <div className="error-state">
          <div className="error-icon">⚠️</div>
          <div className="error-message">Failed to load keys</div>
          <div className="error-details">{error.message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="page keys">
      <div className="page-header">
        <div className="page-header-content">
          <div>
            <h1 className="page-title">Cryptographic Keys</h1>
            <p className="page-subtitle">Manage your cryptographic keys for signing and encryption</p>
          </div>
          <button 
            className="primary-button"
            onClick={() => setShowGenerateModal(true)}
            disabled={generateKey.isPending}
          >
            {generateKey.isPending ? 'Generating...' : '+ Generate Key'}
          </button>
        </div>
      </div>

      <div className="page-content">
        {keys.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">🔑</div>
            <div className="empty-title">No keys found</div>
            <div className="empty-description">
              Create your first cryptographic key to get started
            </div>
            <button 
              className="empty-action"
              onClick={() => setShowGenerateModal(true)}
            >
              Generate Key
            </button>
          </div>
        ) : (
          <div className="keys-grid">
            {keys.map((key: any) => (
              <KeyCard
                key={key.id}
                keyData={key}
                onDelete={() => handleDeleteKey(key.id, key.id)}
                isDeleting={deleteKey.isPending}
              />
            ))}
          </div>
        )}
      </div>

      {showGenerateModal && (
        <KeyGenerationModal
          onGenerate={handleGenerateKey}
          onClose={() => setShowGenerateModal(false)}
          isGenerating={generateKey.isPending}
        />
      )}
    </div>
  );
}