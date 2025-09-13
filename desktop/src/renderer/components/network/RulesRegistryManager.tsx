import React, { useState } from 'react';
import { 
  useActiveRules, 
  useRulesUpdates 
} from '../../hooks/useWalletAPI';
import { useNotification } from '../../contexts/NotificationContext';

interface Rule {
  id: string;
  name: string;
  version: string;
  description: string;
  type: 'scoring' | 'validation' | 'governance' | 'adjudication';
  status: 'active' | 'pending' | 'deprecated';
  validFrom: string;
  validTo?: string;
  timeLockDays: number;
  author: string;
  hash: string;
  signature: string;
  parameters: Record<string, any>;
  dependencies: string[];
  conflicts: string[];
  lastUpdated: string;
}

interface RulesUpdate {
  id: string;
  type: 'proposed' | 'approved' | 'activated' | 'deprecated';
  ruleId: string;
  ruleName: string;
  version: string;
  timestamp: string;
  proposer: string;
  votes?: {
    for: number;
    against: number;
    abstain: number;
  };
  activationDate?: string;
  description: string;
}

export function RulesRegistryManager() {
  const [selectedRule, setSelectedRule] = useState<string | null>(null);
  const [filterType, setFilterType] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [showUpdateDetails, setShowUpdateDetails] = useState<string | null>(null);
  
  const { data: activeRules = [], isLoading: rulesLoading, refetch: refetchRules } = useActiveRules();
  const { data: rulesUpdates = [], isLoading: updatesLoading } = useRulesUpdates();
  const { showNotification } = useNotification();

  const getRuleTypeIcon = (type: string): string => {
    switch (type) {
      case 'scoring': return '📊';
      case 'validation': return '✅';
      case 'governance': return '🏛️';
      case 'adjudication': return '⚖️';
      default: return '📝';
    }
  };

  const getRuleStatusColor = (status: string): string => {
    switch (status) {
      case 'active': return '#22c55e';
      case 'pending': return '#f59e0b';
      case 'deprecated': return '#ef4444';
      default: return '#6b7280';
    }
  };

  const getUpdateTypeColor = (type: string): string => {
    switch (type) {
      case 'proposed': return '#3b82f6';
      case 'approved': return '#f59e0b';
      case 'activated': return '#22c55e';
      case 'deprecated': return '#ef4444';
      default: return '#6b7280';
    }
  };

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  const isRuleExpiring = (rule: Rule): boolean => {
    if (!rule.validTo) return false;
    const expiryDate = new Date(rule.validTo);
    const now = new Date();
    const daysUntilExpiry = (expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24);
    return daysUntilExpiry <= 30; // Warn if expiring within 30 days
  };

  const canActivateRule = (rule: Rule): boolean => {
    const activationDate = new Date(rule.validFrom);
    const now = new Date();
    return activationDate <= now && rule.status === 'pending';
  };

  const filteredRules = activeRules.filter((rule: Rule) => {
    if (filterType !== 'all' && rule.type !== filterType) return false;
    if (filterStatus !== 'all' && rule.status !== filterStatus) return false;
    return true;
  });

  const renderRuleDetails = (rule: Rule) => (
    <div className="rule-details-panel">
      <div className="details-header">
        <h4>{rule.name}</h4>
        <div className="rule-meta">
          <span className="rule-version">v{rule.version}</span>
          <span 
            className="rule-status"
            style={{ color: getRuleStatusColor(rule.status) }}
          >
            {rule.status.toUpperCase()}
          </span>
        </div>
        <button 
          onClick={() => setSelectedRule(null)}
          className="close-details"
        >
          ×
        </button>
      </div>
      
      <div className="details-content">
        <div className="detail-section">
          <h5>Rule Information</h5>
          <div className="detail-item">
            <span className="label">ID:</span>
            <span className="value monospace">{rule.id}</span>
          </div>
          <div className="detail-item">
            <span className="label">Type:</span>
            <span className="value">
              {getRuleTypeIcon(rule.type)} {rule.type}
            </span>
          </div>
          <div className="detail-item">
            <span className="label">Description:</span>
            <span className="value">{rule.description}</span>
          </div>
          <div className="detail-item">
            <span className="label">Author:</span>
            <span className="value">{rule.author}</span>
          </div>
        </div>

        <div className="detail-section">
          <h5>Validity</h5>
          <div className="detail-item">
            <span className="label">Valid From:</span>
            <span className="value">{formatTimestamp(rule.validFrom)}</span>
          </div>
          {rule.validTo && (
            <div className="detail-item">
              <span className="label">Valid To:</span>
              <span 
                className={`value ${isRuleExpiring(rule) ? 'warning' : ''}`}
              >
                {formatTimestamp(rule.validTo)}
                {isRuleExpiring(rule) && ' ⚠️ Expiring Soon'}
              </span>
            </div>
          )}
          <div className="detail-item">
            <span className="label">Time Lock:</span>
            <span className="value">{rule.timeLockDays} days</span>
          </div>
        </div>

        <div className="detail-section">
          <h5>Cryptographic Verification</h5>
          <div className="detail-item">
            <span className="label">Hash:</span>
            <span className="value monospace">{rule.hash}</span>
          </div>
          <div className="detail-item">
            <span className="label">Signature:</span>
            <span className="value monospace">{rule.signature.substring(0, 32)}...</span>
          </div>
        </div>

        {Object.keys(rule.parameters).length > 0 && (
          <div className="detail-section">
            <h5>Parameters</h5>
            <div className="parameters-list">
              {Object.entries(rule.parameters).map(([key, value]) => (
                <div key={key} className="parameter-item">
                  <span className="parameter-key">{key}:</span>
                  <span className="parameter-value">
                    {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {rule.dependencies.length > 0 && (
          <div className="detail-section">
            <h5>Dependencies</h5>
            <div className="dependencies-list">
              {rule.dependencies.map((dep, index) => (
                <span key={index} className="dependency-tag">
                  {dep}
                </span>
              ))}
            </div>
          </div>
        )}

        {rule.conflicts.length > 0 && (
          <div className="detail-section">
            <h5>Conflicts</h5>
            <div className="conflicts-list">
              {rule.conflicts.map((conflict, index) => (
                <span key={index} className="conflict-tag">
                  {conflict}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="rule-actions">
        {canActivateRule(rule) && (
          <button className="action-button activate">
            Activate Rule
          </button>
        )}
        {rule.status === 'active' && (
          <button className="action-button deprecate">
            Propose Deprecation
          </button>
        )}
        <button className="action-button export">
          Export Rule
        </button>
      </div>
    </div>
  );

  const renderRulesList = () => (
    <div className="rules-list">
      <div className="list-header">
        <div className="list-title">
          <h4>Active Rules ({filteredRules.length})</h4>
        </div>
        
        <div className="list-filters">
          <div className="filter-group">
            <label>Type:</label>
            <select 
              value={filterType} 
              onChange={(e) => setFilterType(e.target.value)}
            >
              <option value="all">All Types</option>
              <option value="scoring">Scoring</option>
              <option value="validation">Validation</option>
              <option value="governance">Governance</option>
              <option value="adjudication">Adjudication</option>
            </select>
          </div>
          
          <div className="filter-group">
            <label>Status:</label>
            <select 
              value={filterStatus} 
              onChange={(e) => setFilterStatus(e.target.value)}
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="pending">Pending</option>
              <option value="deprecated">Deprecated</option>
            </select>
          </div>
          
          <button 
            onClick={() => refetchRules()}
            disabled={rulesLoading}
            className="refresh-button"
          >
            {rulesLoading ? '⟳' : '↻'} Refresh
          </button>
        </div>
      </div>
      
      <div className="rules-table">
        <div className="table-header">
          <div className="col-name">Rule Name</div>
          <div className="col-type">Type</div>
          <div className="col-version">Version</div>
          <div className="col-status">Status</div>
          <div className="col-validity">Validity</div>
          <div className="col-actions">Actions</div>
        </div>
        
        <div className="table-body">
          {filteredRules.map((rule: Rule) => (
            <div key={rule.id} className="table-row">
              <div className="col-name">
                <div className="rule-name-info">
                  <div className="rule-name">{rule.name}</div>
                  <div className="rule-description">{rule.description}</div>
                </div>
              </div>
              
              <div className="col-type">
                <span className="type-badge">
                  {getRuleTypeIcon(rule.type)} {rule.type}
                </span>
              </div>
              
              <div className="col-version">
                v{rule.version}
              </div>
              
              <div className="col-status">
                <span 
                  className="status-badge"
                  style={{ color: getRuleStatusColor(rule.status) }}
                >
                  {rule.status}
                </span>
                {isRuleExpiring(rule) && (
                  <span className="expiry-warning" title="Expiring soon">⚠️</span>
                )}
              </div>
              
              <div className="col-validity">
                <div className="validity-info">
                  <div>From: {new Date(rule.validFrom).toLocaleDateString()}</div>
                  {rule.validTo && (
                    <div>To: {new Date(rule.validTo).toLocaleDateString()}</div>
                  )}
                </div>
              </div>
              
              <div className="col-actions">
                <button
                  onClick={() => setSelectedRule(rule.id)}
                  className="action-button details"
                  title="View details"
                >
                  👁
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );

  const renderUpdatesStream = () => (
    <div className="rules-updates">
      <div className="updates-header">
        <h4>Recent Rule Updates ({rulesUpdates.length})</h4>
      </div>
      
      <div className="updates-list">
        {rulesUpdates.length === 0 ? (
          <div className="no-updates">No recent rule updates</div>
        ) : (
          rulesUpdates.map((update: RulesUpdate) => (
            <div key={update.id} className="update-item">
              <div className="update-header">
                <span 
                  className="update-type"
                  style={{ color: getUpdateTypeColor(update.type) }}
                >
                  {update.type.toUpperCase()}
                </span>
                <span className="update-time">
                  {formatTimestamp(update.timestamp)}
                </span>
              </div>
              
              <div className="update-content">
                <div className="update-title">
                  {update.ruleName} v{update.version}
                </div>
                <div className="update-description">
                  {update.description}
                </div>
                
                {update.votes && (
                  <div className="voting-info">
                    <span className="vote-count for">👍 {update.votes.for}</span>
                    <span className="vote-count against">👎 {update.votes.against}</span>
                    <span className="vote-count abstain">🤷 {update.votes.abstain}</span>
                  </div>
                )}
                
                <div className="update-proposer">
                  Proposed by: {update.proposer}
                </div>
              </div>
              
              <div className="update-actions">
                <button
                  onClick={() => setShowUpdateDetails(update.id)}
                  className="details-button"
                >
                  View Details
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );

  if (rulesLoading && updatesLoading) {
    return (
      <div className="rules-registry-manager loading">
        <div className="loading-spinner">Loading rules registry...</div>
      </div>
    );
  }

  return (
    <div className="rules-registry-manager">
      <div className="manager-header">
        <h3>Rules Registry Manager</h3>
        <p className="description">
          Monitor active rules, track updates, and manage governance decisions.
        </p>
      </div>

      <div className="manager-content">
        <div className="content-grid">
          <div className="rules-section">
            {renderRulesList()}
          </div>
          
          <div className="updates-section">
            {renderUpdatesStream()}
          </div>
        </div>
        
        {selectedRule && (
          <div className="rule-details-overlay">
            {renderRuleDetails(activeRules.find((r: Rule) => r.id === selectedRule)!)}
          </div>
        )}
      </div>
    </div>
  );
}

export default RulesRegistryManager;