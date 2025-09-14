import React, { useState } from 'react';

interface TrustScoreSearchProps {
  onSearch: (filters: {
    did: string;
    context: string;
    minScore?: number;
    maxScore?: number;
    dateFrom?: string;
    dateTo?: string;
  }) => void;
}

export function TrustScoreSearch({ onSearch }: TrustScoreSearchProps): JSX.Element {
  const [searchDID, setSearchDID] = useState('');
  const [searchContext, setSearchContext] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [minScore, setMinScore] = useState<number | undefined>();
  const [maxScore, setMaxScore] = useState<number | undefined>();
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const [isSearching, setIsSearching] = useState(false);

  const contexts = [
    { value: '', label: 'All Contexts' },
    { value: 'general', label: 'General' },
    { value: 'commerce', label: 'Commerce' },
    { value: 'hiring', label: 'Hiring' },
    { value: 'social', label: 'Social' },
    { value: 'financial', label: 'Financial' },
  ];

  const handleSearch = async () => {
    if (!searchDID.trim()) return;
    
    setIsSearching(true);
    try {
      await onSearch({
        did: searchDID.trim(),
        context: searchContext,
        minScore: minScore,
        maxScore: maxScore,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
      });
    } finally {
      setIsSearching(false);
    }
  };

  const handleKeyPress = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter') {
      handleSearch();
    }
  };

  const clearSearch = () => {
    setSearchDID('');
    setSearchContext('');
    setMinScore(undefined);
    setMaxScore(undefined);
    setDateFrom('');
    setDateTo('');
    onSearch({
      did: '',
      context: '',
      minScore: undefined,
      maxScore: undefined,
      dateFrom: undefined,
      dateTo: undefined,
    });
  };

  const isValidDID = (did: string) => {
    return did.startsWith('did:') && did.length > 10;
  };

  const handleDIDChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setSearchDID(value);
    
    // Auto-search if valid DID format is detected
    if (isValidDID(value) && value.length > 20) {
      setTimeout(() => {
        if (searchDID === value) {
          handleSearch();
        }
      }, 500);
    }
  };

  return (
    <div className="trust-score-search">
      <div className="search-header">
        <div className="search-title">
          <span className="search-icon">🔍</span>
          <span>Search Trust Scores</span>
        </div>
      </div>

      <div className="search-form">
        <div className="search-row">
          <div className="form-group search-did-group">
            <label className="form-label">DID (Decentralized Identifier)</label>
            <div className="search-input-wrapper">
              <input
                type="text"
                className={`form-input search-input ${searchDID && !isValidDID(searchDID) ? 'invalid' : ''}`}
                value={searchDID}
                onChange={handleDIDChange}
                onKeyPress={handleKeyPress}
                placeholder="did:key:z6Mk... or did:web:example.com"
              />
              {searchDID && (
                <button
                  className="clear-button"
                  onClick={() => setSearchDID('')}
                  title="Clear DID"
                >
                  ×
                </button>
              )}
            </div>
            {searchDID && !isValidDID(searchDID) && (
              <div className="form-help error">
                Please enter a valid DID starting with "did:"
              </div>
            )}
            <div className="form-help">
              Enter the DID of the entity whose trust score you want to view
            </div>
          </div>

          <div className="form-group search-context-group">
            <label className="form-label">Context</label>
            <select
              className="form-input context-select"
              value={searchContext}
              onChange={(e) => setSearchContext(e.target.value)}
            >
              {contexts.map((context) => (
                <option key={context.value} value={context.value}>
                  {context.label}
                </option>
              ))}
            </select>
            <div className="form-help">
              Filter scores by specific context or view all
            </div>
          </div>
        </div>

        <div className="search-actions">
          <button
            className="filter-toggle"
            onClick={() => setShowAdvanced(!showAdvanced)}
            type="button"
          >
            {showAdvanced ? '▲' : '▼'} Advanced Filters
          </button>
          
          <button
            className="primary-button search-button"
            onClick={handleSearch}
            disabled={!searchDID.trim() || !isValidDID(searchDID) || isSearching}
          >
            {isSearching ? (
              <>
                <span className="spinner-small"></span>
                Searching...
              </>
            ) : (
              <>
                🔍 Search Trust Score
              </>
            )}
          </button>
          
          {searchDID && (
            <button
              className="secondary-button"
              onClick={clearSearch}
              disabled={isSearching}
            >
              Clear
            </button>
          )}
        </div>

        {showAdvanced && (
          <div className="advanced-filters">
            <div className="filter-row">
              <div className="filter-group">
                <label className="form-label">Score Range</label>
                <div className="score-range-inputs">
                  <input
                    type="number"
                    className="form-input score-input"
                    value={minScore || ''}
                    onChange={(e) => setMinScore(e.target.value ? parseInt(e.target.value) : undefined)}
                    placeholder="Min"
                    min="0"
                    max="100"
                  />
                  <span className="range-separator">-</span>
                  <input
                    type="number"
                    className="form-input score-input"
                    value={maxScore || ''}
                    onChange={(e) => setMaxScore(e.target.value ? parseInt(e.target.value) : undefined)}
                    placeholder="Max"
                    min="0"
                    max="100"
                  />
                </div>
                <div className="form-help">Filter by trust score range (0-100)</div>
              </div>
            </div>

            <div className="filter-row">
              <div className="filter-group">
                <label className="form-label">Date Range</label>
                <div className="date-range-inputs">
                  <input
                    type="date"
                    className="form-input date-input"
                    value={dateFrom}
                    onChange={(e) => setDateFrom(e.target.value)}
                  />
                  <span className="range-separator">to</span>
                  <input
                    type="date"
                    className="form-input date-input"
                    value={dateTo}
                    onChange={(e) => setDateTo(e.target.value)}
                  />
                </div>
                <div className="form-help">Filter by score creation/update date</div>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="search-examples">
        <div className="examples-title">Example DIDs:</div>
        <div className="example-dids">
          <button
            className="example-did"
            onClick={() => setSearchDID('did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK')}
            title="Click to use this example"
          >
            did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK
          </button>
          <button
            className="example-did"
            onClick={() => setSearchDID('did:web:example.com')}
            title="Click to use this example"
          >
            did:web:example.com
          </button>
        </div>
      </div>

      <div className="search-info">
        <div className="info-content">
          <div className="info-icon">💡</div>
          <div className="info-text">
            <div className="info-title">About Trust Scores</div>
            <div className="info-description">
              Trust scores are calculated based on vouches, reports, and other trust events. 
              Scores range from 0-100 and can be context-specific (e.g., hiring, commerce).
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}