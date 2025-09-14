import React, { useState } from 'react';
import { useTrustScores, useTrustScore, useDIDs } from '../hooks/useWalletAPI';
import { TrustScoreCard } from '../components/trust-scores/TrustScoreCard';
import { TrustScoreSearch } from '../components/trust-scores/TrustScoreSearch';

export function TrustScores(): JSX.Element {
  const [selectedDID, setSelectedDID] = useState<string>('');
  const [selectedContext, setSelectedContext] = useState<string>('');
  const [searchDID, setSearchDID] = useState<string>('');
  
  const { data: dids = [] } = useDIDs();
  const { data: allScores = [], isLoading: scoresLoading, error } = useTrustScores();
  const { data: specificScore, isLoading: specificLoading, refetch: refetchSpecific } = useTrustScore(
    searchDID || selectedDID,
    selectedContext
  );

  const contexts = [
    { value: '', label: 'All Contexts' },
    { value: 'general', label: 'General' },
    { value: 'commerce', label: 'Commerce' },
    { value: 'hiring', label: 'Hiring' },
  ];

  const handleSearch = async (filters: {
    did: string;
    context: string;
    minScore?: number;
    maxScore?: number;
    dateFrom?: string;
    dateTo?: string;
  }) => {
    setSearchDID(filters.did);
    setSelectedContext(filters.context);
    if (filters.did) {
      refetchSpecific();
    }
    // TODO: Implement additional filtering with minScore, maxScore, dateFrom, dateTo
    console.log('Advanced filters:', filters);
  };

  if (scoresLoading) {
    return (
      <div className="page trust-scores">
        <div className="page-header">
          <h1 className="page-title">Trust Scores</h1>
          <p className="page-subtitle">View trust scores and reputation data</p>
        </div>
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading trust scores...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="page trust-scores">
      <div className="page-header">
        <div className="page-header-content">
          <div>
            <h1 className="page-title">Trust Scores</h1>
            <p className="page-subtitle">View trust scores and reputation data</p>
          </div>
        </div>
      </div>

      <div className="page-content">
        <div className="trust-scores-controls">
          <TrustScoreSearch onSearch={handleSearch} />
          
          <div className="context-filter">
            <label className="filter-label">Filter by Context:</label>
            <select 
              className="context-select"
              value={selectedContext}
              onChange={(e) => setSelectedContext(e.target.value)}
            >
              {contexts.map((context) => (
                <option key={context.value} value={context.value}>
                  {context.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {searchDID && specificScore && !specificLoading ? (
          <div className="search-results">
            <div className="results-header">
              <h3>Trust Score for {searchDID}</h3>
              {selectedContext && <span className="context-badge">{selectedContext}</span>}
            </div>
            <TrustScoreCard scoreData={specificScore} />
          </div>
        ) : searchDID && specificLoading ? (
          <div className="search-results">
            <div className="loading-state">
              <div className="spinner"></div>
              <p>Loading trust score...</p>
            </div>
          </div>
        ) : null}

        <div className="my-scores-section">
          <h3 className="section-title">My Trust Scores</h3>
          {dids.length === 0 ? (
            <div className="info-notice">
              <div className="notice-icon">🆔</div>
              <div className="notice-content">
                <div className="notice-title">No DIDs Available</div>
                <div className="notice-text">
                  Create a DID first to start building your trust reputation.
                </div>
              </div>
            </div>
          ) : (
            <div className="my-scores-grid">
              {dids.map((did: any) => (
                <div key={did.did} className="my-score-item">
                  <div className="did-info">
                    <div className="did-label">Your DID:</div>
                    <div className="did-value">{did.did}</div>
                  </div>
                  <button
                    className="view-score-button"
                    onClick={() => handleSearch({ did: did.did, context: '' })}
                  >
                    View Trust Score
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {allScores.length === 0 && !searchDID ? (
          <div className="empty-state">
            <div className="empty-icon">⭐</div>
            <div className="empty-title">No trust scores available</div>
            <div className="empty-description">
              Trust scores will appear here once trust events are created and processed
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}