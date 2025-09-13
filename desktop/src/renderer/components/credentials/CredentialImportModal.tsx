import React, { useState } from 'react';

interface CredentialImportModalProps {
  onImport: (credential: any, metadata: any) => void;
  onClose: () => void;
  isImporting: boolean;
}

export function CredentialImportModal({ onImport, onClose, isImporting }: CredentialImportModalProps): JSX.Element {
  const [importMethod, setImportMethod] = useState<'paste' | 'file'>('paste');
  const [credentialText, setCredentialText] = useState('');
  const [metadata, setMetadata] = useState({
    tags: '',
    description: '',
  });

  const handleImport = () => {
    try {
      const credential = JSON.parse(credentialText);
      const credentialMetadata = {
        tags: metadata.tags ? metadata.tags.split(',').map(t => t.trim()) : [],
        description: metadata.description || '',
        importedAt: new Date().toISOString(),
      };
      
      onImport(credential, credentialMetadata);
    } catch (error) {
      alert('Invalid JSON format. Please check your credential data.');
    }
  };

  const handleFileImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setCredentialText(e.target?.result as string);
      };
      reader.readAsText(file);
    }
  };

  const isValidJSON = () => {
    try {
      JSON.parse(credentialText);
      return credentialText.trim().length > 0;
    } catch {
      return false;
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Import Verifiable Credential</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="form-group">
            <label className="form-label">Import Method</label>
            <div className="import-method-tabs">
              <button
                className={`method-tab ${importMethod === 'paste' ? 'active' : ''}`}
                onClick={() => setImportMethod('paste')}
              >
                📋 Paste JSON
              </button>
              <button
                className={`method-tab ${importMethod === 'file' ? 'active' : ''}`}
                onClick={() => setImportMethod('file')}
              >
                📁 Upload File
              </button>
            </div>
          </div>

          {importMethod === 'paste' ? (
            <div className="form-group">
              <label className="form-label">Credential JSON</label>
              <textarea
                className="credential-textarea"
                value={credentialText}
                onChange={(e) => setCredentialText(e.target.value)}
                placeholder="Paste your verifiable credential JSON here..."
                rows={10}
              />
              <div className="form-help">
                Paste the complete JSON representation of your verifiable credential
              </div>
            </div>
          ) : (
            <div className="form-group">
              <label className="form-label">Select File</label>
              <input
                type="file"
                accept=".json,.txt"
                onChange={handleFileImport}
                className="file-input"
              />
              <div className="form-help">
                Choose a JSON file containing your verifiable credential
              </div>
            </div>
          )}

          <div className="form-group">
            <label className="form-label">Tags (Optional)</label>
            <input
              type="text"
              className="form-input"
              value={metadata.tags}
              onChange={(e) => setMetadata(prev => ({ ...prev, tags: e.target.value }))}
              placeholder="education, work, certification"
            />
            <div className="form-help">
              Comma-separated tags to help organize your credentials
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Description (Optional)</label>
            <textarea
              className="form-textarea"
              value={metadata.description}
              onChange={(e) => setMetadata(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Personal notes about this credential..."
              rows={3}
            />
          </div>

          <div className="info-notice">
            <div className="notice-icon">ℹ️</div>
            <div className="notice-content">
              <div className="notice-title">About Verifiable Credentials</div>
              <div className="notice-text">
                Verifiable credentials are digital certificates that can be cryptographically verified. 
                They contain claims about a subject and can be shared while maintaining privacy.
              </div>
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <button 
            className="secondary-button" 
            onClick={onClose}
            disabled={isImporting}
          >
            Cancel
          </button>
          <button 
            className="primary-button" 
            onClick={handleImport}
            disabled={isImporting || !isValidJSON()}
          >
            {isImporting ? 'Importing...' : 'Import Credential'}
          </button>
        </div>
      </div>
    </div>
  );
}