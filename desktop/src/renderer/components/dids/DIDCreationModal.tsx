import React, { useState } from 'react';

interface DIDCreationModalProps {
  keys: any[];
  onCreate: (keyId: string, method: string) => void;
  onClose: () => void;
  isCreating: boolean;
}

export function DIDCreationModal({ keys, onCreate, onClose, isCreating }: DIDCreationModalProps): JSX.Element {
  const [selectedKeyId, setSelectedKeyId] = useState(keys.length > 0 ? keys[0].id : '');
  const [selectedMethod, setSelectedMethod] = useState('key');

  const didMethods = [
    {
      value: 'key',
      label: 'did:key',
      description: 'Cryptographically derived DID from public key',
      icon: '🔐',
    },
  ];

  const handleCreate = () => {
    if (selectedKeyId && selectedMethod) {
      onCreate(selectedKeyId, selectedMethod);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Create New DID</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="form-group">
            <label className="form-label">Select Key</label>
            <select 
              className="form-select"
              value={selectedKeyId}
              onChange={(e) => setSelectedKeyId(e.target.value)}
            >
              {keys.map((key) => (
                <option key={key.id} value={key.id}>
                  {key.id} ({key.keyType})
                </option>
              ))}
            </select>
            <div className="form-help">
              The DID will be cryptographically tied to this key
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">DID Method</label>
            <div className="method-options">
              {didMethods.map((method) => (
                <div
                  key={method.value}
                  className={`method-option ${selectedMethod === method.value ? 'selected' : ''}`}
                  onClick={() => setSelectedMethod(method.value)}
                >
                  <div className="method-icon">{method.icon}</div>
                  <div className="method-info">
                    <div className="method-label">{method.label}</div>
                    <div className="method-description">{method.description}</div>
                  </div>
                  <div className="method-radio">
                    <input
                      type="radio"
                      name="method"
                      value={method.value}
                      checked={selectedMethod === method.value}
                      onChange={(e) => setSelectedMethod(e.target.value)}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="info-notice">
            <div className="notice-icon">ℹ️</div>
            <div className="notice-content">
              <div className="notice-title">About DIDs</div>
              <div className="notice-text">
                A Decentralized Identifier (DID) is a globally unique identifier that enables verifiable, 
                self-sovereign digital identity. Your DID will be derived from your selected key.
              </div>
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <button 
            className="secondary-button" 
            onClick={onClose}
            disabled={isCreating}
          >
            Cancel
          </button>
          <button 
            className="primary-button" 
            onClick={handleCreate}
            disabled={isCreating || !selectedKeyId}
          >
            {isCreating ? 'Creating...' : 'Create DID'}
          </button>
        </div>
      </div>
    </div>
  );
}