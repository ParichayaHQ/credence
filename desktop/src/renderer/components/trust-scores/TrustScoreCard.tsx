import React, { useState } from 'react';

interface TrustScoreCardProps {
  scoreData: any;
}

export function TrustScoreCard({ scoreData }: TrustScoreCardProps): JSX.Element {
  const [showDetails, setShowDetails] = useState(false);

  const getScoreColor = (score: number) => {
    if (score >= 80) return 'excellent';
    if (score >= 60) return 'good';
    if (score >= 40) return 'fair';
    return 'poor';
  };

  const getScoreIcon = (score: number) => {
    if (score >= 80) return '⭐';
    if (score >= 60) return '✅';
    if (score >= 40) return '🔶';
    return '⚠️';
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text).then(() => {
      console.log(`${label} copied to clipboard`);
    });
  };

  const score = scoreData.score || 0;
  const context = scoreData.context || 'general';
  const lastUpdated = scoreData.lastUpdated || scoreData.created || new Date().toISOString();

  return (
    <div className="trust-score-card">
      <div className="trust-score-header">
        <div className="score-info">
          <div className="score-icon">
            {getScoreIcon(score)}
          </div>
          <div className="score-details">
            <div className="score-value">
              <span className={`score-number ${getScoreColor(score)}`}>
                {score}
              </span>
              <span className="score-max">/100</span>
            </div>
            <div className="score-context">
              Context: {context}
            </div>
          </div>
        </div>
        <div className="score-actions">
          <div className={`score-badge ${getScoreColor(score)}`}>
            {score >= 80 ? 'Excellent' : score >= 60 ? 'Good' : score >= 40 ? 'Fair' : 'Poor'}
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

      <div className="trust-score-body">
        <div className="score-meta">
          <div className="score-meta-item">
            <span className="meta-label">DID:</span>
            <span className="meta-value subject-id">
              {scoreData.subjectDid || scoreData.did || 'Unknown'}
            </span>
          </div>
          <div className="score-meta-item">
            <span className="meta-label">Last Updated:</span>
            <span className="meta-value">{formatDate(lastUpdated)}</span>
          </div>
          {scoreData.eventCount && (
            <div className="score-meta-item">
              <span className="meta-label">Based on:</span>
              <span className="meta-value">{scoreData.eventCount} events</span>
            </div>
          )}
        </div>

        <div className="score-actions-bar">
          <button
            className="action-button-small"
            onClick={() => copyToClipboard(scoreData.subjectDid || scoreData.did || '', 'DID')}
            title="Copy DID"
          >
            📋 Copy DID
          </button>
        </div>

        {showDetails && (
          <div className="score-details-expanded">
            <div className="score-section">
              <div className="section-title">Score Breakdown</div>
              <div className="score-breakdown">
                {scoreData.breakdown ? (
                  <div className="breakdown-list">
                    {Object.entries(scoreData.breakdown).map(([key, value]: [string, any]) => (
                      <div key={key} className="breakdown-item">
                        <span className="breakdown-key">{key}:</span>
                        <span className="breakdown-value">{value}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="breakdown-simple">
                    <div className="breakdown-item">
                      <span className="breakdown-key">Overall Score:</span>
                      <span className="breakdown-value">{score}/100</span>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {scoreData.recentEvents && scoreData.recentEvents.length > 0 && (
              <div className="score-section">
                <div className="section-title">Recent Events</div>
                <div className="recent-events">
                  {scoreData.recentEvents.slice(0, 5).map((event: any, index: number) => (
                    <div key={index} className="event-item">
                      <div className="event-type">
                        {event.type === 'vouch' ? '👍' : '👎'} {event.type}
                      </div>
                      <div className="event-details">
                        <div className="event-date">{formatDate(event.created)}</div>
                        {event.context && <div className="event-context">Context: {event.context}</div>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="score-section">
              <div className="section-title">Full Data</div>
              <div className="score-json">
                <pre className="json-content">
                  {JSON.stringify(scoreData, null, 2)}
                </pre>
                <button
                  className="copy-button"
                  onClick={() => copyToClipboard(
                    JSON.stringify(scoreData, null, 2),
                    'Trust Score Data'
                  )}
                >
                  📋 Copy JSON
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}