import React, { useState, useEffect, useRef } from 'react';
import { useEventStream } from '../../hooks/useWalletAPI';

interface RealtimeEvent {
  id: string;
  type: 'vouch' | 'report' | 'revocation' | 'checkpoint' | 'rule-update';
  from?: string;
  to?: string;
  context?: string;
  timestamp: string;
  epoch?: string;
  payloadCID?: string;
  networkLatency?: number;
  confidence?: number;
  details?: any;
}

interface EventFeedProps {
  maxEvents?: number;
  autoScroll?: boolean;
  showFilters?: boolean;
}

export function RealTimeEventFeed({ 
  maxEvents = 50, 
  autoScroll = true, 
  showFilters = true 
}: EventFeedProps) {
  const [events, setEvents] = useState<RealtimeEvent[]>([]);
  const [filters, setFilters] = useState({
    types: [] as string[],
    contexts: [] as string[],
    showOnlyMyEvents: false
  });
  const [isPaused, setIsPaused] = useState(false);
  
  const feedRef = useRef<HTMLDivElement>(null);
  const { data: eventStream = [] } = useEventStream();

  // Simulate real-time events (in production, this would be WebSocket/SSE)
  useEffect(() => {
    if (!isPaused && eventStream.length > 0) {
      setEvents(prev => {
        const newEvents = eventStream
          .filter((event: RealtimeEvent) => !prev.find(e => e.id === event.id))
          .slice(0, maxEvents - prev.length);
        
        const combined = [...newEvents, ...prev].slice(0, maxEvents);
        return combined;
      });
    }
  }, [eventStream, isPaused, maxEvents]);

  // Auto-scroll to top when new events arrive
  useEffect(() => {
    if (autoScroll && !isPaused && feedRef.current) {
      feedRef.current.scrollTop = 0;
    }
  }, [events, autoScroll, isPaused]);

  const getEventIcon = (type: string): string => {
    switch (type) {
      case 'vouch': return '👍';
      case 'report': return '⚠️';
      case 'revocation': return '🚫';
      case 'checkpoint': return '📋';
      case 'rule-update': return '⚙️';
      default: return '📄';
    }
  };

  const getEventColor = (type: string): string => {
    switch (type) {
      case 'vouch': return '#22c55e';
      case 'report': return '#f59e0b';
      case 'revocation': return '#ef4444';
      case 'checkpoint': return '#3b82f6';
      case 'rule-update': return '#8b5cf6';
      default: return '#6b7280';
    }
  };

  const formatTimestamp = (timestamp: string): string => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    
    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`;
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  };

  const formatDID = (did: string): string => {
    if (!did) return 'Unknown';
    if (did.startsWith('did:key:')) {
      const key = did.replace('did:key:', '');
      return `${key.substring(0, 8)}...${key.substring(key.length - 4)}`;
    }
    return did;
  };

  const filteredEvents = events.filter(event => {
    if (filters.types.length > 0 && !filters.types.includes(event.type)) return false;
    if (filters.contexts.length > 0 && event.context && !filters.contexts.includes(event.context)) return false;
    // TODO: Implement showOnlyMyEvents filter
    return true;
  });

  const availableTypes = Array.from(new Set(events.map(e => e.type)));
  const availableContexts = Array.from(new Set(events.map(e => e.context).filter(Boolean)));

  const renderFilters = () => (
    <div className="event-filters">
      <div className="filter-group">
        <label>Event Types:</label>
        <div className="checkbox-group">
          {availableTypes.map(type => (
            <label key={type} className="checkbox-label">
              <input
                type="checkbox"
                checked={filters.types.includes(type)}
                onChange={(e) => {
                  if (e.target.checked) {
                    setFilters(prev => ({ ...prev, types: [...prev.types, type] }));
                  } else {
                    setFilters(prev => ({ ...prev, types: prev.types.filter(t => t !== type) }));
                  }
                }}
              />
              <span className="checkbox-text">
                {getEventIcon(type)} {type}
              </span>
            </label>
          ))}
        </div>
      </div>

      {availableContexts.length > 0 && (
        <div className="filter-group">
          <label>Contexts:</label>
          <div className="checkbox-group">
            {availableContexts.map(context => (
              <label key={context} className="checkbox-label">
                <input
                  type="checkbox"
                  checked={filters.contexts.includes(context!)}
                  onChange={(e) => {
                    if (e.target.checked) {
                      setFilters(prev => ({ ...prev, contexts: [...prev.contexts, context!] }));
                    } else {
                      setFilters(prev => ({ ...prev, contexts: prev.contexts.filter(c => c !== context) }));
                    }
                  }}
                />
                <span className="checkbox-text">{context}</span>
              </label>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  const renderEvent = (event: RealtimeEvent) => (
    <div key={event.id} className="event-item">
      <div className="event-header">
        <div className="event-type" style={{ color: getEventColor(event.type) }}>
          <span className="event-icon">{getEventIcon(event.type)}</span>
          <span className="event-type-text">{event.type.toUpperCase()}</span>
        </div>
        <div className="event-timestamp">
          {formatTimestamp(event.timestamp)}
        </div>
      </div>

      <div className="event-content">
        {event.from && event.to && (
          <div className="event-participants">
            <span className="participant">
              <span className="label">From:</span>
              <span className="value">{formatDID(event.from)}</span>
            </span>
            <span className="arrow">→</span>
            <span className="participant">
              <span className="label">To:</span>
              <span className="value">{formatDID(event.to)}</span>
            </span>
          </div>
        )}

        {event.context && (
          <div className="event-context">
            <span className="label">Context:</span>
            <span className="value context-badge">{event.context}</span>
          </div>
        )}

        {event.epoch && (
          <div className="event-epoch">
            <span className="label">Epoch:</span>
            <span className="value">{event.epoch}</span>
          </div>
        )}

        <div className="event-metadata">
          {event.networkLatency && (
            <span className="metadata-item">
              <span className="label">Latency:</span>
              <span className="value">{event.networkLatency}ms</span>
            </span>
          )}
          {event.confidence && (
            <span className="metadata-item">
              <span className="label">Confidence:</span>
              <span className="value">{(event.confidence * 100).toFixed(1)}%</span>
            </span>
          )}
        </div>

        {event.payloadCID && (
          <div className="event-payload">
            <span className="label">Payload:</span>
            <span className="value payload-cid">{event.payloadCID}</span>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div className="realtime-event-feed">
      <div className="feed-header">
        <div className="feed-title">
          <h3>Live Event Feed</h3>
          <div className="feed-status">
            <span className={`status-indicator ${isPaused ? 'paused' : 'live'}`}>
              {isPaused ? '⏸' : '🔴'} {isPaused ? 'Paused' : 'Live'}
            </span>
            <span className="event-count">
              {filteredEvents.length} events
            </span>
          </div>
        </div>

        <div className="feed-controls">
          <button
            onClick={() => setIsPaused(!isPaused)}
            className={`control-button ${isPaused ? 'play' : 'pause'}`}
          >
            {isPaused ? '▶️ Resume' : '⏸️ Pause'}
          </button>
          
          <button
            onClick={() => setEvents([])}
            className="control-button clear"
          >
            🗑️ Clear
          </button>

          {showFilters && (
            <button
              onClick={() => setFilters({ types: [], contexts: [], showOnlyMyEvents: false })}
              className="control-button clear-filters"
              disabled={filters.types.length === 0 && filters.contexts.length === 0}
            >
              Clear Filters
            </button>
          )}
        </div>
      </div>

      {showFilters && availableTypes.length > 0 && renderFilters()}

      <div className="feed-content" ref={feedRef}>
        {filteredEvents.length === 0 ? (
          <div className="empty-feed">
            <div className="empty-icon">📡</div>
            <div className="empty-text">
              {isPaused ? 'Event feed is paused' : 'Waiting for events...'}
            </div>
            {events.length > 0 && filteredEvents.length === 0 && (
              <div className="filter-hint">
                All events are filtered out. Adjust your filters to see events.
              </div>
            )}
          </div>
        ) : (
          <div className="events-list">
            {filteredEvents.map(renderEvent)}
          </div>
        )}
      </div>
    </div>
  );
}

export default RealTimeEventFeed;