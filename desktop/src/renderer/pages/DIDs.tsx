import React, { useState } from 'react';
import { useDIDs, useKeys, useCreateDID } from '../hooks/useWalletAPI';
import { DIDCreationModal } from '../components/dids/DIDCreationModal';
import { DIDCard } from '../components/dids/DIDCard';

export function DIDs(): JSX.Element {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const { data: dids = [], isLoading: didsLoading, error } = useDIDs();
  const { data: keys = [] } = useKeys();
  const createDID = useCreateDID();

  const handleCreateDID = async (keyId: string, method: string) => {
    await createDID.mutateAsync({ keyId, method });
    setShowCreateModal(false);
  };

  if (didsLoading) {
    return (
      <div className="page dids">
        <div className="page-header">
          <h1 className="page-title">Decentralized Identifiers</h1>
          <p className="page-subtitle">Manage your DIDs for identity and verification</p>
        </div>
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading DIDs...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page dids">
        <div className="page-header">
          <h1 className="page-title">Decentralized Identifiers</h1>
          <p className="page-subtitle">Manage your DIDs for identity and verification</p>
        </div>
        <div className="error-state">
          <div className="error-icon">⚠️</div>
          <div className="error-message">Failed to load DIDs</div>
          <div className="error-details">{error.message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="page dids">
      <div className="page-header">
        <div className="page-header-content">
          <div>
            <h1 className="page-title">Decentralized Identifiers</h1>
            <p className="page-subtitle">Manage your DIDs for identity and verification</p>
          </div>
          <button 
            className="primary-button"
            onClick={() => setShowCreateModal(true)}
            disabled={createDID.isPending || keys.length === 0}
          >
            {createDID.isPending ? 'Creating...' : '+ Create DID'}
          </button>
        </div>
      </div>

      <div className="page-content">
        {keys.length === 0 && dids.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">🔑</div>
            <div className="empty-title">No keys available</div>
            <div className="empty-description">
              You need to generate a cryptographic key before creating a DID
            </div>
            <button 
              className="empty-action"
              onClick={() => window.location.hash = '#/keys'}
            >
              Generate Key
            </button>
          </div>
        ) : dids.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">🆔</div>
            <div className="empty-title">No DIDs found</div>
            <div className="empty-description">
              Create your first DID to establish your digital identity
            </div>
            <button 
              className="empty-action"
              onClick={() => setShowCreateModal(true)}
            >
              Create DID
            </button>
          </div>
        ) : (
          <div className="dids-grid">
            {dids.map((did: any) => (
              <DIDCard
                key={did.did}
                didData={did}
              />
            ))}
          </div>
        )}
      </div>

      {showCreateModal && (
        <DIDCreationModal
          keys={keys}
          onCreate={handleCreateDID}
          onClose={() => setShowCreateModal(false)}
          isCreating={createDID.isPending}
        />
      )}
    </div>
  );
}