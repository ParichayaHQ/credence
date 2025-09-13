import React, { useState } from 'react';
import { useCredentials, useStoreCredential, useDeleteCredential } from '../hooks/useWalletAPI';
import { CredentialImportModal } from '../components/credentials/CredentialImportModal';
import { CredentialCard } from '../components/credentials/CredentialCard';

export function Credentials(): JSX.Element {
  const [showImportModal, setShowImportModal] = useState(false);
  const [selectedCredentials, setSelectedCredentials] = useState<Set<string>>(new Set());
  const [isInBatchMode, setIsInBatchMode] = useState(false);
  const { data: credentials = [], isLoading, error } = useCredentials();
  const storeCredential = useStoreCredential();
  const deleteCredential = useDeleteCredential();

  const handleImportCredential = async (credential: any, metadata: any) => {
    await storeCredential.mutateAsync({ credential, metadata });
    setShowImportModal(false);
  };

  const handleDeleteCredential = async (credentialId: string, credentialName: string) => {
    if (confirm(`Are you sure you want to delete the credential "${credentialName}"? This action cannot be undone.`)) {
      await deleteCredential.mutateAsync(credentialId);
    }
  };

  const exportAllCredentials = () => {
    const exportData = {
      exportedAt: new Date().toISOString(),
      credentialCount: credentials.length,
      credentials: credentials
    };
    
    const dataStr = JSON.stringify(exportData, null, 2);
    const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
    
    const exportFileDefaultName = `credentials_backup_${new Date().toISOString().split('T')[0]}.json`;
    
    const linkElement = document.createElement('a');
    linkElement.setAttribute('href', dataUri);
    linkElement.setAttribute('download', exportFileDefaultName);
    linkElement.click();
    linkElement.remove();
  };

  const handleBatchToggle = () => {
    setIsInBatchMode(!isInBatchMode);
    setSelectedCredentials(new Set());
  };

  const handleCredentialSelect = (credentialId: string, selected: boolean) => {
    const newSelected = new Set(selectedCredentials);
    if (selected) {
      newSelected.add(credentialId);
    } else {
      newSelected.delete(credentialId);
    }
    setSelectedCredentials(newSelected);
  };

  const handleSelectAll = () => {
    if (selectedCredentials.size === credentials.length) {
      setSelectedCredentials(new Set());
    } else {
      setSelectedCredentials(new Set(credentials.map((c: any) => c.id)));
    }
  };

  const handleBatchExport = () => {
    const selectedData = credentials.filter((c: any) => selectedCredentials.has(c.id));
    const exportData = {
      exportedAt: new Date().toISOString(),
      credentialCount: selectedData.length,
      credentials: selectedData
    };
    
    const dataStr = JSON.stringify(exportData, null, 2);
    const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
    
    const exportFileDefaultName = `credentials_selected_${selectedCredentials.size}_${new Date().toISOString().split('T')[0]}.json`;
    
    const linkElement = document.createElement('a');
    linkElement.setAttribute('href', dataUri);
    linkElement.setAttribute('download', exportFileDefaultName);
    linkElement.click();
    linkElement.remove();
  };

  const handleBatchDelete = async () => {
    const selectedCount = selectedCredentials.size;
    const confirmation = confirm(
      `Are you sure you want to delete ${selectedCount} selected credential${selectedCount > 1 ? 's' : ''}? This action cannot be undone.`
    );
    
    if (!confirmation) return;

    try {
      const deletePromises = Array.from(selectedCredentials).map(id =>
        deleteCredential.mutateAsync(id)
      );
      await Promise.all(deletePromises);
      setSelectedCredentials(new Set());
      setIsInBatchMode(false);
    } catch (error) {
      console.error('Batch delete failed:', error);
      alert('Some credentials could not be deleted. Please try again.');
    }
  };

  if (isLoading) {
    return (
      <div className="page credentials">
        <div className="page-header">
          <h1 className="page-title">Verifiable Credentials</h1>
          <p className="page-subtitle">Store and manage your verifiable credentials</p>
        </div>
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading credentials...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page credentials">
        <div className="page-header">
          <h1 className="page-title">Verifiable Credentials</h1>
          <p className="page-subtitle">Store and manage your verifiable credentials</p>
        </div>
        <div className="error-state">
          <div className="error-icon">⚠️</div>
          <div className="error-message">Failed to load credentials</div>
          <div className="error-details">{error.message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="page credentials">
      <div className="page-header">
        <div className="page-header-content">
          <div>
            <h1 className="page-title">Verifiable Credentials</h1>
            <p className="page-subtitle">Store and manage your verifiable credentials</p>
          </div>
          <div className="header-actions">
            {!isInBatchMode ? (
              <>
                <button 
                  className="tertiary-button"
                  onClick={handleBatchToggle}
                  disabled={credentials.length === 0}
                >
                  ☑️ Select Multiple
                </button>
                <button 
                  className="secondary-button"
                  onClick={exportAllCredentials}
                  disabled={credentials.length === 0}
                >
                  📤 Export All
                </button>
                <button 
                  className="primary-button"
                  onClick={() => setShowImportModal(true)}
                  disabled={storeCredential.isPending}
                >
                  {storeCredential.isPending ? 'Importing...' : '+ Import Credential'}
                </button>
              </>
            ) : (
              <>
                <div className="batch-info">
                  {selectedCredentials.size} of {credentials.length} selected
                </div>
                <button 
                  className="tertiary-button"
                  onClick={handleSelectAll}
                >
                  {selectedCredentials.size === credentials.length ? 'Deselect All' : 'Select All'}
                </button>
                <button 
                  className="secondary-button"
                  onClick={handleBatchExport}
                  disabled={selectedCredentials.size === 0}
                >
                  📤 Export Selected
                </button>
                <button 
                  className="danger-button"
                  onClick={handleBatchDelete}
                  disabled={selectedCredentials.size === 0 || deleteCredential.isPending}
                >
                  🗑️ Delete Selected
                </button>
                <button 
                  className="tertiary-button"
                  onClick={handleBatchToggle}
                >
                  Cancel
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      <div className="page-content">
        {credentials.length === 0 ? (
          <div className="empty-state">
            <div className="empty-icon">📜</div>
            <div className="empty-title">No credentials found</div>
            <div className="empty-description">
              Import or receive your first verifiable credential
            </div>
            <button 
              className="empty-action"
              onClick={() => setShowImportModal(true)}
            >
              Import Credential
            </button>
          </div>
        ) : (
          <div className="credentials-grid">
            {credentials.map((credential: any) => (
              <CredentialCard
                key={credential.id}
                credentialData={credential}
                onDelete={() => handleDeleteCredential(credential.id, credential.type?.[0] || 'Unknown')}
                isDeleting={deleteCredential.isPending}
                isInBatchMode={isInBatchMode}
                isSelected={selectedCredentials.has(credential.id)}
                onSelect={(selected) => handleCredentialSelect(credential.id, selected)}
              />
            ))}
          </div>
        )}
      </div>

      {showImportModal && (
        <CredentialImportModal
          onImport={handleImportCredential}
          onClose={() => setShowImportModal(false)}
          isImporting={storeCredential.isPending}
        />
      )}
    </div>
  );
}