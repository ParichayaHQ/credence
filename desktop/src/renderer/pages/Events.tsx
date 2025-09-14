import React, { useState } from 'react';
import { useEvents, useCreateEvent, useDIDs } from '../hooks/useWalletAPI';
import { EventCard } from '../components/events/EventCard';
import { CreateEventModal } from '../components/events/CreateEventModal';

export function Events(): JSX.Element {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [eventFilter, setEventFilter] = useState<'all' | 'vouches' | 'reports'>('all');
  const [contextFilter, setContextFilter] = useState<string>('');

  const { data: events = [], isLoading, error } = useEvents();
  const { data: dids = [] } = useDIDs();
  const createEvent = useCreateEvent();

  const contexts = [
    { value: '', label: 'All Contexts' },
    { value: 'general', label: 'General' },
    { value: 'commerce', label: 'Commerce' },
    { value: 'hiring', label: 'Hiring' },
    { value: 'social', label: 'Social' },
    { value: 'financial', label: 'Financial' },
  ];

  const handleCreateEvent = async (eventData: {
    type: 'vouch' | 'report';
    targetDid: string;
    context: string;
    description: string;
    rating?: number;
    evidence?: string[];
    issuerDid: string;
  }) => {
    // Transform the eventData to match the API expected format
    const apiEventData = {
      type: eventData.type,
      from: eventData.issuerDid,
      to: eventData.targetDid,
      context: eventData.context,
      payloadCID: undefined, // Optional payload field
    };
    
    await createEvent.mutateAsync(apiEventData);
    setShowCreateModal(false);
  };

  const filteredEvents = events.filter((event: any) => {
    if (eventFilter !== 'all' && event.type !== eventFilter.slice(0, -1)) return false;
    if (contextFilter && event.context !== contextFilter) return false;
    return true;
  });

  const eventStats = {
    total: events.length,
    vouches: events.filter((e: any) => e.type === 'vouch').length,
    reports: events.filter((e: any) => e.type === 'report').length,
  };

  if (isLoading) {
    return (
      <div className="page events">
        <div className="page-header">
          <h1 className="page-title">Trust Events</h1>
          <p className="page-subtitle">Manage vouches and reports for building trust</p>
        </div>
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading events...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page events">
        <div className="page-header">
          <h1 className="page-title">Trust Events</h1>
          <p className="page-subtitle">Manage vouches and reports for building trust</p>
        </div>
        <div className="error-state">
          <div className="error-icon">⚠️</div>
          <div className="error-message">Failed to load events</div>
          <div className="error-details">{error.message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="page events">
      <div className="page-header">
        <div className="page-header-content">
          <div>
            <h1 className="page-title">Trust Events</h1>
            <p className="page-subtitle">Manage vouches and reports for building trust</p>
          </div>
          <button 
            className="primary-button"
            onClick={() => setShowCreateModal(true)}
            disabled={createEvent.isPending || dids.length === 0}
          >
            {createEvent.isPending ? 'Creating...' : '+ Create Event'}
          </button>
        </div>
      </div>

      <div className="page-content">
        {dids.length === 0 ? (
          <div className="info-notice">
            <div className="notice-icon">🆔</div>
            <div className="notice-content">
              <div className="notice-title">No DIDs Available</div>
              <div className="notice-text">
                Create a DID first to start creating trust events.
              </div>
            </div>
          </div>
        ) : (
          <>
            <div className="events-stats">
              <div className="stat-card">
                <div className="stat-value">{eventStats.total}</div>
                <div className="stat-label">Total Events</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">{eventStats.vouches}</div>
                <div className="stat-label">Vouches</div>
                <div className="stat-icon">👍</div>
              </div>
              <div className="stat-card">
                <div className="stat-value">{eventStats.reports}</div>
                <div className="stat-label">Reports</div>
                <div className="stat-icon">👎</div>
              </div>
            </div>

            <div className="events-filters">
              <div className="filter-group">
                <label className="filter-label">Event Type:</label>
                <div className="filter-tabs">
                  <button
                    className={`filter-tab ${eventFilter === 'all' ? 'active' : ''}`}
                    onClick={() => setEventFilter('all')}
                  >
                    All
                  </button>
                  <button
                    className={`filter-tab ${eventFilter === 'vouches' ? 'active' : ''}`}
                    onClick={() => setEventFilter('vouches')}
                  >
                    👍 Vouches
                  </button>
                  <button
                    className={`filter-tab ${eventFilter === 'reports' ? 'active' : ''}`}
                    onClick={() => setEventFilter('reports')}
                  >
                    👎 Reports
                  </button>
                </div>
              </div>

              <div className="filter-group">
                <label className="filter-label">Context:</label>
                <select 
                  className="context-select"
                  value={contextFilter}
                  onChange={(e) => setContextFilter(e.target.value)}
                >
                  {contexts.map((context) => (
                    <option key={context.value} value={context.value}>
                      {context.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {filteredEvents.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">
                  {eventFilter === 'vouches' ? '👍' : eventFilter === 'reports' ? '👎' : '📋'}
                </div>
                <div className="empty-title">
                  {eventFilter === 'all' ? 'No events found' : 
                   eventFilter === 'vouches' ? 'No vouches found' : 'No reports found'}
                </div>
                <div className="empty-description">
                  {eventFilter === 'all' 
                    ? 'Create your first trust event to build reputation data'
                    : `Create your first ${eventFilter.slice(0, -1)} to get started`
                  }
                </div>
                <button 
                  className="empty-action"
                  onClick={() => setShowCreateModal(true)}
                >
                  Create Event
                </button>
              </div>
            ) : (
              <div className="events-grid">
                {filteredEvents.map((event: any) => (
                  <EventCard
                    key={event.id}
                    eventData={event}
                  />
                ))}
              </div>
            )}
          </>
        )}
      </div>

      {showCreateModal && (
        <CreateEventModal
          onCreateEvent={handleCreateEvent}
          onClose={() => setShowCreateModal(false)}
          isCreating={createEvent.isPending}
          availableDIDs={dids}
        />
      )}
    </div>
  );
}