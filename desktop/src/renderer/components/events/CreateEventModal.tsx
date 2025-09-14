import React, { useState } from 'react';

interface CreateEventModalProps {
  onCreateEvent: (eventData: {
    type: 'vouch' | 'report';
    targetDid: string;
    context: string;
    description: string;
    rating?: number;
    evidence?: string[];
    issuerDid: string;
  }) => void;
  onClose: () => void;
  isCreating: boolean;
  availableDIDs: any[];
}

export function CreateEventModal({ onCreateEvent, onClose, isCreating, availableDIDs }: CreateEventModalProps): JSX.Element {
  const [eventType, setEventType] = useState<'vouch' | 'report'>('vouch');
  const [targetDid, setTargetDid] = useState('');
  const [issuerDid, setIssuerDid] = useState(availableDIDs[0]?.did || '');
  const [context, setContext] = useState('general');
  const [description, setDescription] = useState('');
  const [rating, setRating] = useState(5);
  const [evidence, setEvidence] = useState<string[]>(['']);

  const contexts = [
    { value: 'general', label: 'General' },
    { value: 'commerce', label: 'Commerce' },
    { value: 'hiring', label: 'Hiring' },
    { value: 'social', label: 'Social' },
    { value: 'financial', label: 'Financial' },
  ];

  const handleCreate = () => {
    const eventData = {
      type: eventType,
      targetDid: targetDid.trim(),
      context,
      description: description.trim(),
      rating,
      evidence: evidence.filter(e => e.trim() !== ''),
      issuerDid,
    };
    
    onCreateEvent(eventData);
  };

  const addEvidenceField = () => {
    setEvidence([...evidence, '']);
  };

  const removeEvidenceField = (index: number) => {
    if (evidence.length > 1) {
      setEvidence(evidence.filter((_, i) => i !== index));
    }
  };

  const updateEvidenceField = (index: number, value: string) => {
    const newEvidence = [...evidence];
    newEvidence[index] = value;
    setEvidence(newEvidence);
  };

  const isValidDID = (did: string) => {
    return did.startsWith('did:') && did.length > 10;
  };

  const canCreate = () => {
    return targetDid.trim() && 
           isValidDID(targetDid) && 
           issuerDid && 
           description.trim() && 
           !isCreating;
  };

  const getRatingLabel = () => {
    if (eventType === 'vouch') {
      return 'Rating (1-5 stars)';
    } else {
      return 'Severity (1-5 scale)';
    }
  };

  const getRatingDescription = () => {
    if (eventType === 'vouch') {
      return 'Rate your trust in this entity (5 = highest trust)';
    } else {
      return 'Rate the severity of the issue (5 = most severe)';
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content large-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2 className="modal-title">
            Create Trust {eventType === 'vouch' ? 'Vouch' : 'Report'}
          </h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          <div className="form-group">
            <label className="form-label">Event Type</label>
            <div className="event-type-tabs">
              <button
                className={`type-tab ${eventType === 'vouch' ? 'active' : ''}`}
                onClick={() => setEventType('vouch')}
                disabled={isCreating}
              >
                👍 Vouch
                <div className="tab-description">Positive trust endorsement</div>
              </button>
              <button
                className={`type-tab ${eventType === 'report' ? 'active' : ''}`}
                onClick={() => setEventType('report')}
                disabled={isCreating}
              >
                👎 Report
                <div className="tab-description">Negative trust report</div>
              </button>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label className="form-label">Target DID *</label>
              <input
                type="text"
                className={`form-input ${targetDid && !isValidDID(targetDid) ? 'invalid' : ''}`}
                value={targetDid}
                onChange={(e) => setTargetDid(e.target.value)}
                placeholder="did:key:z6Mk... or did:web:example.com"
                disabled={isCreating}
              />
              {targetDid && !isValidDID(targetDid) && (
                <div className="form-help error">
                  Please enter a valid DID starting with "did:"
                </div>
              )}
              <div className="form-help">
                The DID of the entity you are {eventType === 'vouch' ? 'vouching for' : 'reporting'}
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">Your DID *</label>
              <select
                className="form-input"
                value={issuerDid}
                onChange={(e) => setIssuerDid(e.target.value)}
                disabled={isCreating}
              >
                {availableDIDs.map((did: any) => (
                  <option key={did.did} value={did.did}>
                    {did.did}
                  </option>
                ))}
              </select>
              <div className="form-help">
                The DID that will issue this {eventType}
              </div>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label className="form-label">Context</label>
              <select
                className="form-input"
                value={context}
                onChange={(e) => setContext(e.target.value)}
                disabled={isCreating}
              >
                {contexts.map((ctx) => (
                  <option key={ctx.value} value={ctx.value}>
                    {ctx.label}
                  </option>
                ))}
              </select>
              <div className="form-help">
                The context or domain for this trust event
              </div>
            </div>

            <div className="form-group">
              <label className="form-label">{getRatingLabel()} *</label>
              <div className="rating-input">
                <input
                  type="range"
                  min="1"
                  max="5"
                  value={rating}
                  onChange={(e) => setRating(parseInt(e.target.value))}
                  className="rating-slider"
                  disabled={isCreating}
                />
                <div className="rating-value">{rating}/5</div>
              </div>
              <div className="form-help">
                {getRatingDescription()}
              </div>
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Description *</label>
            <textarea
              className="form-textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={eventType === 'vouch' ? 
                "Describe why you trust this entity..." : 
                "Describe the issue or concern..."
              }
              rows={4}
              disabled={isCreating}
            />
            <div className="form-help">
              Provide details about your {eventType === 'vouch' ? 'positive experience' : 'concerns'}
            </div>
          </div>

          <div className="form-group">
            <label className="form-label">Evidence (Optional)</label>
            <div className="evidence-fields">
              {evidence.map((item, index) => (
                <div key={index} className="evidence-field">
                  <input
                    type="text"
                    className="form-input"
                    value={item}
                    onChange={(e) => updateEvidenceField(index, e.target.value)}
                    placeholder={`Evidence item ${index + 1}`}
                    disabled={isCreating}
                  />
                  {evidence.length > 1 && (
                    <button
                      className="remove-evidence-button"
                      onClick={() => removeEvidenceField(index)}
                      disabled={isCreating}
                    >
                      ×
                    </button>
                  )}
                </div>
              ))}
              <button
                className="add-evidence-button"
                onClick={addEvidenceField}
                disabled={isCreating}
              >
                + Add Evidence
              </button>
            </div>
            <div className="form-help">
              Links, transaction IDs, or other supporting information
            </div>
          </div>

          <div className="info-notice">
            <div className="notice-icon">💡</div>
            <div className="notice-content">
              <div className="notice-title">About Trust Events</div>
              <div className="notice-text">
                {eventType === 'vouch' 
                  ? "Vouches are positive endorsements that increase trust scores. Use them to recognize good behavior, successful transactions, or trustworthy interactions."
                  : "Reports are negative feedback that can decrease trust scores. Use them responsibly to flag problematic behavior, scams, or untrustworthy actions."
                }
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
            className={`primary-button ${eventType === 'report' ? 'danger' : ''}`}
            onClick={handleCreate}
            disabled={!canCreate()}
          >
            {isCreating ? (
              <>Creating {eventType}...</>
            ) : (
              <>Create {eventType === 'vouch' ? 'Vouch' : 'Report'}</>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}