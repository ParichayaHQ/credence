import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  useKeys, 
  useDIDs, 
  useCredentials, 
  useEvents, 
  useTrustScores 
} from '../../hooks/useWalletAPI';

interface SearchResult {
  id: string;
  type: 'key' | 'did' | 'credential' | 'event' | 'trust-score' | 'page' | 'command';
  title: string;
  subtitle: string;
  description?: string;
  icon: string;
  action: () => void;
  score: number;
  category: string;
  metadata?: Record<string, any>;
}

interface GlobalSearchProps {
  isOpen: boolean;
  onClose: () => void;
  placeholder?: string;
  maxResults?: number;
}

export function GlobalSearch({ 
  isOpen, 
  onClose, 
  placeholder = "Search keys, DIDs, credentials, events...", 
  maxResults = 20 
}: GlobalSearchProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  
  const inputRef = useRef<HTMLInputElement>(null);
  const resultsRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  // Data hooks
  const { data: keys = [] } = useKeys();
  const { data: dids = [] } = useDIDs();
  const { data: credentials = [] } = useCredentials();
  const { data: events = [] } = useEvents();
  const { data: trustScores = [] } = useTrustScores();

  // Pages and commands for navigation
  const pages = [
    { id: 'dashboard', title: 'Dashboard', path: '/dashboard', icon: '📊', keywords: ['dashboard', 'home', 'overview'] },
    { id: 'keys', title: 'Keys', path: '/keys', icon: '🔑', keywords: ['keys', 'keypair', 'cryptographic'] },
    { id: 'dids', title: 'DIDs', path: '/dids', icon: '🆔', keywords: ['dids', 'identity', 'decentralized'] },
    { id: 'credentials', title: 'Credentials', path: '/credentials', icon: '📜', keywords: ['credentials', 'vc', 'verifiable'] },
    { id: 'events', title: 'Events', path: '/events', icon: '📝', keywords: ['events', 'vouch', 'report'] },
    { id: 'trust-scores', title: 'Trust Scores', path: '/trust-scores', icon: '⭐', keywords: ['trust', 'score', 'reputation'] },
    { id: 'network', title: 'Network', path: '/network', icon: '🌐', keywords: ['network', 'p2p', 'peers', 'checkpoints'] },
    { id: 'settings', title: 'Settings', path: '/settings', icon: '⚙️', keywords: ['settings', 'preferences', 'configuration'] },
  ];

  const commands = [
    { id: 'lock', title: 'Lock Wallet', action: () => { /* Lock wallet */ }, icon: '🔒', keywords: ['lock', 'secure'] },
    { id: 'backup', title: 'Create Backup', action: () => { /* Create backup */ }, icon: '💾', keywords: ['backup', 'export'] },
    { id: 'generate-key', title: 'Generate New Key', action: () => navigate('/keys'), icon: '🔑', keywords: ['generate', 'new', 'key'] },
    { id: 'create-did', title: 'Create New DID', action: () => navigate('/dids'), icon: '🆔', keywords: ['create', 'new', 'did'] },
    { id: 'import-credential', title: 'Import Credential', action: () => navigate('/credentials'), icon: '📥', keywords: ['import', 'credential'] },
    { id: 'create-vouch', title: 'Create Vouch', action: () => navigate('/events'), icon: '👍', keywords: ['vouch', 'create', 'endorse'] },
    { id: 'create-report', title: 'Create Report', action: () => navigate('/events'), icon: '⚠️', keywords: ['report', 'create', 'flag'] },
  ];

  // Search function
  const performSearch = useMemo(() => {
    return (searchQuery: string): SearchResult[] => {
      if (!searchQuery.trim()) return [];

      const query = searchQuery.toLowerCase().trim();
      const allResults: SearchResult[] = [];

      // Search function to calculate relevance score
      const calculateScore = (item: any, fields: string[], keywords: string[] = []): number => {
        let score = 0;
        const searchTerms = query.split(' ');

        // Check exact matches in primary fields
        fields.forEach((field, index) => {
          const value = (item[field] || '').toLowerCase();
          searchTerms.forEach(term => {
            if (value.includes(term)) {
              score += (10 - index) * (value === term ? 10 : value.startsWith(term) ? 8 : 5);
            }
          });
        });

        // Check keyword matches
        keywords.forEach(keyword => {
          searchTerms.forEach(term => {
            if (keyword.toLowerCase().includes(term)) {
              score += 3;
            }
          });
        });

        return score;
      };

      // Search pages
      pages.forEach(page => {
        const score = calculateScore(page, ['title'], page.keywords);
        if (score > 0) {
          allResults.push({
            id: `page-${page.id}`,
            type: 'page',
            title: page.title,
            subtitle: 'Navigate to page',
            icon: page.icon,
            action: () => {
              navigate(page.path);
              onClose();
            },
            score,
            category: 'Pages'
          });
        }
      });

      // Search commands
      commands.forEach(command => {
        const score = calculateScore(command, ['title'], command.keywords);
        if (score > 0) {
          allResults.push({
            id: `command-${command.id}`,
            type: 'command',
            title: command.title,
            subtitle: 'Execute command',
            icon: command.icon,
            action: () => {
              command.action();
              onClose();
            },
            score,
            category: 'Commands'
          });
        }
      });

      // Search keys
      keys.forEach(key => {
        const score = calculateScore(key, ['type', 'keyId', 'algorithm'], ['key', 'crypto', 'keypair']);
        if (score > 0) {
          allResults.push({
            id: `key-${key.keyId}`,
            type: 'key',
            title: `${key.type} Key`,
            subtitle: key.keyId.substring(0, 16) + '...',
            description: `${key.algorithm} key created ${new Date(key.createdAt).toLocaleDateString()}`,
            icon: '🔑',
            action: () => {
              navigate(`/keys?highlight=${key.keyId}`);
              onClose();
            },
            score,
            category: 'Keys',
            metadata: key
          });
        }
      });

      // Search DIDs
      dids.forEach(did => {
        const score = calculateScore(did, ['did', 'method'], ['did', 'identity', 'decentralized']);
        if (score > 0) {
          allResults.push({
            id: `did-${did.did}`,
            type: 'did',
            title: 'DID Identity',
            subtitle: did.did.length > 50 ? did.did.substring(0, 50) + '...' : did.did,
            description: `${did.method} method created ${new Date(did.createdAt).toLocaleDateString()}`,
            icon: '🆔',
            action: () => {
              navigate(`/dids?highlight=${encodeURIComponent(did.did)}`);
              onClose();
            },
            score,
            category: 'DIDs',
            metadata: did
          });
        }
      });

      // Search credentials
      credentials.forEach(credential => {
        const issuerName = credential.issuer?.name || 'Unknown Issuer';
        const credentialType = credential.type?.[credential.type.length - 1] || 'Credential';
        
        const score = calculateScore(
          { ...credential, issuerName, credentialType },
          ['credentialId', 'issuerName', 'credentialType'],
          ['credential', 'vc', 'verifiable']
        );
        
        if (score > 0) {
          allResults.push({
            id: `credential-${credential.credentialId}`,
            type: 'credential',
            title: credentialType,
            subtitle: `Issued by ${issuerName}`,
            description: `Created ${new Date(credential.issuanceDate).toLocaleDateString()}`,
            icon: '📜',
            action: () => {
              navigate(`/credentials?highlight=${credential.credentialId}`);
              onClose();
            },
            score,
            category: 'Credentials',
            metadata: credential
          });
        }
      });

      // Search events
      events.forEach(event => {
        const fromDID = event.from.length > 20 ? event.from.substring(0, 20) + '...' : event.from;
        const toDID = event.to.length > 20 ? event.to.substring(0, 20) + '...' : event.to;
        
        const score = calculateScore(
          { ...event, fromDID, toDID },
          ['type', 'context', 'from', 'to'],
          ['event', 'vouch', 'report']
        );
        
        if (score > 0) {
          allResults.push({
            id: `event-${event.eventId}`,
            type: 'event',
            title: `${event.type.toUpperCase()} Event`,
            subtitle: `${fromDID} → ${toDID}`,
            description: `Context: ${event.context} • ${new Date(event.timestamp).toLocaleDateString()}`,
            icon: event.type === 'vouch' ? '👍' : '⚠️',
            action: () => {
              navigate(`/events?highlight=${event.eventId}`);
              onClose();
            },
            score,
            category: 'Events',
            metadata: event
          });
        }
      });

      // Search trust scores
      trustScores.forEach(score => {
        const didDisplay = score.did.length > 30 ? score.did.substring(0, 30) + '...' : score.did;
        
        const searchScore = calculateScore(
          { ...score, didDisplay },
          ['did', 'context'],
          ['trust', 'score', 'reputation']
        );
        
        if (searchScore > 0) {
          allResults.push({
            id: `score-${score.did}-${score.context}`,
            type: 'trust-score',
            title: `Trust Score (${score.context})`,
            subtitle: didDisplay,
            description: `Score: ${score.score} • Updated ${new Date(score.updatedAt).toLocaleDateString()}`,
            icon: '⭐',
            action: () => {
              navigate(`/trust-scores?highlight=${encodeURIComponent(score.did)}&context=${score.context}`);
              onClose();
            },
            score: searchScore,
            category: 'Trust Scores',
            metadata: score
          });
        }
      });

      // Sort by score and limit results
      return allResults
        .sort((a, b) => b.score - a.score)
        .slice(0, maxResults);
    };
  }, [keys, dids, credentials, events, trustScores, navigate, onClose, maxResults]);

  // Update results when query changes
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      setSelectedIndex(0);
      return;
    }

    setIsLoading(true);
    
    // Debounce search
    const timeoutId = setTimeout(() => {
      const searchResults = performSearch(query);
      setResults(searchResults);
      setSelectedIndex(0);
      setIsLoading(false);
    }, 200);

    return () => clearTimeout(timeoutId);
  }, [query, performSearch]);

  // Handle keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => Math.min(prev + 1, results.length - 1));
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => Math.max(prev - 1, 0));
          break;
        case 'Enter':
          e.preventDefault();
          if (results[selectedIndex]) {
            results[selectedIndex].action();
          }
          break;
        case 'Escape':
          e.preventDefault();
          onClose();
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, results, selectedIndex, onClose]);

  // Focus input when opened
  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  // Scroll selected item into view
  useEffect(() => {
    if (resultsRef.current) {
      const selectedElement = resultsRef.current.querySelector(`[data-index="${selectedIndex}"]`);
      if (selectedElement) {
        selectedElement.scrollIntoView({ block: 'nearest' });
      }
    }
  }, [selectedIndex]);

  if (!isOpen) return null;

  const groupedResults = results.reduce((groups, result) => {
    const category = result.category;
    if (!groups[category]) groups[category] = [];
    groups[category].push(result);
    return groups;
  }, {} as Record<string, SearchResult[]>);

  return (
    <div className="global-search-overlay" onClick={onClose}>
      <div className="global-search-container" onClick={e => e.stopPropagation()}>
        <div className="search-input-container">
          <div className="search-icon">🔍</div>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={placeholder}
            className="search-input"
          />
          <div className="search-shortcut">⌘K</div>
        </div>

        {isLoading && (
          <div className="search-loading">
            <div className="loading-spinner">⟳</div>
            <span>Searching...</span>
          </div>
        )}

        {!isLoading && query && results.length === 0 && (
          <div className="search-no-results">
            <div className="no-results-icon">🔍</div>
            <div className="no-results-text">No results found for "{query}"</div>
            <div className="no-results-hint">
              Try searching for keys, DIDs, credentials, events, or pages
            </div>
          </div>
        )}

        {!isLoading && results.length > 0 && (
          <div className="search-results" ref={resultsRef}>
            {Object.entries(groupedResults).map(([category, categoryResults]) => (
              <div key={category} className="results-category">
                <div className="category-header">{category}</div>
                <div className="category-results">
                  {categoryResults.map((result, categoryIndex) => {
                    const globalIndex = results.indexOf(result);
                    return (
                      <div
                        key={result.id}
                        data-index={globalIndex}
                        className={`search-result ${globalIndex === selectedIndex ? 'selected' : ''}`}
                        onClick={result.action}
                        onMouseEnter={() => setSelectedIndex(globalIndex)}
                      >
                        <div className="result-icon">{result.icon}</div>
                        <div className="result-content">
                          <div className="result-title">{result.title}</div>
                          <div className="result-subtitle">{result.subtitle}</div>
                          {result.description && (
                            <div className="result-description">{result.description}</div>
                          )}
                        </div>
                        <div className="result-type">{result.type}</div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}

        {!query && (
          <div className="search-tips">
            <div className="tips-header">Search Tips</div>
            <div className="tips-list">
              <div className="tip-item">
                <span className="tip-shortcut">↑↓</span>
                <span className="tip-description">Navigate results</span>
              </div>
              <div className="tip-item">
                <span className="tip-shortcut">⏎</span>
                <span className="tip-description">Select result</span>
              </div>
              <div className="tip-item">
                <span className="tip-shortcut">esc</span>
                <span className="tip-description">Close search</span>
              </div>
            </div>
            <div className="search-examples">
              Try searching: "keys", "credentials", "create vouch", "dashboard"
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Search trigger component for header
interface SearchTriggerProps {
  onOpen: () => void;
}

export function SearchTrigger({ onOpen }: SearchTriggerProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        onOpen();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onOpen]);

  return (
    <button className="search-trigger" onClick={onOpen}>
      <span className="search-icon">🔍</span>
      <span className="search-placeholder">Search...</span>
      <span className="search-shortcut">⌘K</span>
    </button>
  );
}

export default GlobalSearch;