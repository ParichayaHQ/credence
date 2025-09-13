import React, { useState } from 'react';

interface KeyGenerationModalProps {
  onGenerate: (keyType: string) => void;
  onClose: () => void;
  isGenerating: boolean;
}

export function KeyGenerationModal({ onGenerate, onClose, isGenerating }: KeyGenerationModalProps): JSX.Element {
  const [selectedKeyType, setSelectedKeyType] = useState('Ed25519');

  const keyTypes = [
    {
      value: 'Ed25519',
      label: 'Ed25519',
      description: 'Fast, secure signing key (recommended)',
      icon: '🔐',
    },
    {
      value: 'Secp256k1',
      label: 'Secp256k1',
      description: 'Bitcoin-compatible elliptic curve key',
      icon: '₿',
    },
  ];

  const handleGenerate = () => {
    onGenerate(selectedKeyType);
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">Generate New Key</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="form-group">
            <label className="form-label">Key Type</label>
            <div className="key-type-options">
              {keyTypes.map((keyType) => (
                <div
                  key={keyType.value}
                  className={`key-type-option ${selectedKeyType === keyType.value ? 'selected' : ''}`}
                  onClick={() => setSelectedKeyType(keyType.value)}
                >
                  <div className="key-type-icon">{keyType.icon}</div>
                  <div className="key-type-info">
                    <div className="key-type-label">{keyType.label}</div>
                    <div className="key-type-description">{keyType.description}</div>
                  </div>
                  <div className="key-type-radio">
                    <input
                      type="radio"
                      name="keyType"
                      value={keyType.value}
                      checked={selectedKeyType === keyType.value}
                      onChange={(e) => setSelectedKeyType(e.target.value)}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="security-notice">
            <div className="notice-icon">🔒</div>
            <div className="notice-content">
              <div className="notice-title">Security Notice</div>
              <div className="notice-text">
                This key will be stored securely in your local wallet. Keep your wallet password safe to protect access to your keys.
              </div>
            </div>
          </div>
        </div>

        <div className="modal-footer">
          <button 
            className="secondary-button" 
            onClick={onClose}
            disabled={isGenerating}
          >
            Cancel
          </button>
          <button 
            className="primary-button" 
            onClick={handleGenerate}
            disabled={isGenerating}
          >
            {isGenerating ? 'Generating...' : 'Generate Key'}
          </button>
        </div>
      </div>
    </div>
  );
}