import React, { useState } from 'react';

interface EventCardProps {
  eventData: any;
}

export function EventCard({ eventData }: EventCardProps): JSX.Element {
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

  const getEventIcon = (type: string) => {
    return type === 'vouch' ? '👍' : '👎';
  };

  const getEventTypeLabel = (type: string) => {
    return type === 'vouch' ? 'Vouch' : 'Report';
  };

  const getEventColor = (type: string) => {
    return type === 'vouch' ? 'positive' : 'negative';
  };

  const getRatingLabel = (type: string, rating: number) => {
    if (type === 'vouch') {
      return `Rating: ${rating}/5`;
    } else {
      const severityLabels = {
        1: 'Low',
        2: 'Medium',
        3: 'High',
        4: 'Critical',
        5: 'Severe'
      };
      return `Severity: ${severityLabels[rating as keyof typeof severityLabels] || rating}`;
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      console.log(`${label} copied to clipboard`);
    });
  };

  const truncateText = (text: string, maxLength: number) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  return (
    <div className={`event-card ${getEventColor(eventData.type)}`}>
      <div className="event-card-header">
        <div className="event-info">
          <div className="event-icon">
            {getEventIcon(eventData.type)}
          </div>
          <div className="event-details">
            <div className="event-type">
              {getEventTypeLabel(eventData.type)}
            </div>
            <div className="event-context">
              Context: {eventData.context || 'general'}
            </div>
          </div>
        </div>
        <div className="event-actions">
          <div className={`event-badge ${getEventColor(eventData.type)}`}>
            {getRatingLabel(eventData.type, eventData.rating || eventData.severity || 3)}
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

      <div className="event-card-body">
        <div className="event-meta">
          <div className="event-meta-item">
            <span className="meta-label">Target:</span>
            <span className="meta-value target-did">
              {truncateText(eventData.targetDid || 'Unknown', 40)}
            </span>
          </div>
          <div className="event-meta-item">
            <span className="meta-label">Issuer:</span>
            <span className="meta-value issuer-did">
              {truncateText(eventData.issuerDid || eventData.created_by || 'Unknown', 40)}
            </span>
          </div>
          <div className="event-meta-item">
            <span className="meta-label">Created:</span>
            <span className="meta-value">{formatDate(eventData.created || eventData.timestamp)}</span>
          </div>
        </div>

        <div className="event-description">
          <div className="description-text">
            {showDetails 
              ? (eventData.description || 'No description provided')
              : truncateText(eventData.description || 'No description provided', 100)
            }
          </div>
        </div>

        <div className="event-actions-bar">
          <button
            className="action-button-small"
            onClick={() => copyToClipboard(eventData.targetDid || '', 'Target DID')}
            title="Copy target DID"
          >
            📋 Copy Target
          </button>
          <button
            className="action-button-small"
            onClick={() => copyToClipboard(eventData.id || '', 'Event ID')}
            title="Copy event ID"
          >
            🆔 Copy ID
          </button>
        </div>

        {showDetails && (
          <div className="event-details-expanded">
            {eventData.evidence && eventData.evidence.length > 0 && (
              <div className="event-section">
                <div className="section-title">Evidence</div>
                <div className="evidence-list">
                  {eventData.evidence.map((evidence: string, index: number) => (
                    <div key={index} className="evidence-item">
                      <div className="evidence-index">{index + 1}.</div>
                      <div className="evidence-content">{evidence}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="event-section">
              <div className="section-title">Event Data</div>
              <div className="event-json">
                <pre className="json-content">
                  {JSON.stringify(eventData, null, 2)}
                </pre>
                <button
                  className="copy-button"
                  onClick={() => copyToClipboard(
                    JSON.stringify(eventData, null, 2),
                    'Event Data'
                  )}
                >
                  📋 Copy JSON
                </button>
              </div>
            </div>

            <div className="event-section">
              <div className="section-title">Metadata</div>
              <div className="metadata-list">
                <div className="metadata-item">
                  <span className="metadata-key">Event ID:</span>
                  <span className="metadata-value">{eventData.id || 'N/A'}</span>
                </div>
                <div className="metadata-item">
                  <span className="metadata-key">Type:</span>
                  <span className="metadata-value">{eventData.type}</span>
                </div>
                <div className="metadata-item">
                  <span className="metadata-key">Context:</span>
                  <span className="metadata-value">{eventData.context || 'general'}</span>
                </div>
                {eventData.updated && (
                  <div className="metadata-item">
                    <span className="metadata-key">Updated:</span>
                    <span className="metadata-value">{formatDate(eventData.updated)}</span>
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