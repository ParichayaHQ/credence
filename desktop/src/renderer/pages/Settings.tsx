import React, { useState, useEffect } from 'react';
import { useKeys, useDIDs, useCredentials, useEvents } from '../hooks/useWalletAPI';
import { useNotification } from '../contexts/NotificationContext';

interface SettingsState {
  // General settings
  theme: 'light' | 'dark' | 'auto';
  language: string;
  autoStart: boolean;
  minimizeToTray: boolean;
  
  // Security settings
  lockOnMinimize: boolean;
  autoLockMinutes: number;
  enableBiometric: boolean;
  requireAuthForExports: boolean;
  sessionTimeoutMinutes: number;
  enableAppLock: boolean;
  
  // Network settings
  walletServiceUrl: string;
  autoConnectPeers: boolean;
  maxPeers: number;
  enableP2PLogging: boolean;
  networkTimeout: number;
  
  // Privacy settings
  shareAnalytics: boolean;
  enableTelemetry: boolean;
  storeEventHistory: boolean;
  autoBackupInterval: 'never' | 'daily' | 'weekly' | 'monthly';
  
  // Notification settings
  enableNotifications: boolean;
  securityAlerts: boolean;
  eventNotifications: boolean;
  networkUpdates: boolean;
  soundEnabled: boolean;
  
  // Advanced settings
  debugMode: boolean;
  developerTools: boolean;
  customRpcEndpoint: string;
  experimentalFeatures: boolean;
}

export function Settings(): JSX.Element {
  const [activeTab, setActiveTab] = useState('general');
  const [isBackingUp, setIsBackingUp] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  
  const [settings, setSettings] = useState<SettingsState>({
    // General settings
    theme: 'dark',
    language: 'en',
    autoStart: false,
    minimizeToTray: true,
    
    // Security settings
    lockOnMinimize: true,
    autoLockMinutes: 5,
    enableBiometric: false,
    requireAuthForExports: true,
    sessionTimeoutMinutes: 120,
    enableAppLock: false,
    
    // Network settings
    walletServiceUrl: 'http://localhost:8080',
    autoConnectPeers: true,
    maxPeers: 50,
    enableP2PLogging: false,
    networkTimeout: 30,
    
    // Privacy settings
    shareAnalytics: false,
    enableTelemetry: false,
    storeEventHistory: true,
    autoBackupInterval: 'weekly',
    
    // Notification settings
    enableNotifications: true,
    securityAlerts: true,
    eventNotifications: true,
    networkUpdates: false,
    soundEnabled: true,
    
    // Advanced settings
    debugMode: false,
    developerTools: false,
    customRpcEndpoint: '',
    experimentalFeatures: false
  });
  
  const { data: keys = [] } = useKeys();
  const { data: dids = [] } = useDIDs();
  const { data: credentials = [] } = useCredentials();
  const { data: events = [] } = useEvents();
  const { showNotification } = useNotification();

  // Load settings from localStorage on mount
  useEffect(() => {
    const savedSettings = localStorage.getItem('walletSettings');
    if (savedSettings) {
      try {
        setSettings(JSON.parse(savedSettings));
      } catch (error) {
        console.error('Failed to load settings:', error);
      }
    }
  }, []);

  // Save settings to localStorage when they change
  const saveSettings = () => {
    localStorage.setItem('walletSettings', JSON.stringify(settings));
    setHasChanges(false);
    showNotification('Settings saved successfully', 'success');
  };

  const updateSetting = <K extends keyof SettingsState>(key: K, value: SettingsState[K]) => {
    setSettings(prev => ({ ...prev, [key]: value }));
    setHasChanges(true);
  };

  const resetSettings = () => {
    if (confirm('Are you sure you want to reset all settings to their default values?')) {
      setSettings({
        theme: 'dark',
        language: 'en',
        autoStart: false,
        minimizeToTray: true,
        lockOnMinimize: true,
        autoLockMinutes: 5,
        enableBiometric: false,
        requireAuthForExports: true,
        sessionTimeoutMinutes: 120,
        enableAppLock: false,
        walletServiceUrl: 'http://localhost:8080',
        autoConnectPeers: true,
        maxPeers: 50,
        enableP2PLogging: false,
        networkTimeout: 30,
        shareAnalytics: false,
        enableTelemetry: false,
        storeEventHistory: true,
        autoBackupInterval: 'weekly',
        enableNotifications: true,
        securityAlerts: true,
        eventNotifications: true,
        networkUpdates: false,
        soundEnabled: true,
        debugMode: false,
        developerTools: false,
        customRpcEndpoint: '',
        experimentalFeatures: false
      });
      setHasChanges(true);
      showNotification('Settings reset to defaults', 'info');
    }
  };

  const handleFullBackup = async () => {
    setIsBackingUp(true);
    try {
      const backupData = {
        version: '1.0.0',
        timestamp: new Date().toISOString(),
        wallet: {
          keys: keys,
          dids: dids,
          credentials: credentials,
          events: events,
        },
        settings: settings,
        metadata: {
          keyCount: keys.length,
          didCount: dids.length,
          credentialCount: credentials.length,
          eventCount: events.length,
        }
      };

      const dataStr = JSON.stringify(backupData, null, 2);
      const dataUri = 'data:application/json;charset=utf-8,'+ encodeURIComponent(dataStr);
      
      const exportFileDefaultName = `credence_wallet_backup_${new Date().toISOString().split('T')[0]}.json`;
      
      const linkElement = document.createElement('a');
      linkElement.setAttribute('href', dataUri);
      linkElement.setAttribute('download', exportFileDefaultName);
      linkElement.click();
      linkElement.remove();

      showNotification('Backup created successfully', 'success');
    } catch (error) {
      console.error('Backup failed:', error);
      showNotification('Failed to create backup', 'error');
    } finally {
      setIsBackingUp(false);
    }
  };

  const handleRestoreFile = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setIsRestoring(true);
    try {
      const fileContent = await file.text();
      const backupData = JSON.parse(fileContent);

      if (!backupData.wallet || !backupData.version) {
        throw new Error('Invalid backup file format');
      }

      const confirmation = confirm(`This will restore data from ${backupData.timestamp}:\n\n` +
        `• ${backupData.metadata?.keyCount || 0} keys\n` +
        `• ${backupData.metadata?.didCount || 0} DIDs\n` +
        `• ${backupData.metadata?.credentialCount || 0} credentials\n` +
        `• ${backupData.metadata?.eventCount || 0} events\n\n` +
        `This may overwrite existing data. Continue?`);

      if (!confirmation) {
        setIsRestoring(false);
        return;
      }

      // Restore settings if available
      if (backupData.settings) {
        setSettings(backupData.settings);
        setHasChanges(true);
      }

      // TODO: Implement actual restore functionality via API
      console.log('Restore data:', backupData.wallet);
      showNotification('Restore functionality will be implemented with API integration', 'info');
      
    } catch (error) {
      console.error('Restore failed:', error);
      showNotification('Failed to restore from backup', 'error');
    } finally {
      setIsRestoring(false);
      event.target.value = '';
    }
  };

  const tabs = [
    { id: 'general', label: 'General', icon: '⚙️' },
    { id: 'security', label: 'Security', icon: '🔒' },
    { id: 'network', label: 'Network', icon: '🌐' },
    { id: 'privacy', label: 'Privacy', icon: '🛡️' },
    { id: 'notifications', label: 'Notifications', icon: '🔔' },
    { id: 'backup', label: 'Backup & Restore', icon: '💾' },
    { id: 'advanced', label: 'Advanced', icon: '🔧' }
  ];

  const renderGeneralTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>Appearance</h3>
        <div className="setting-item">
          <label>Theme</label>
          <select
            value={settings.theme}
            onChange={(e) => updateSetting('theme', e.target.value as any)}
          >
            <option value="light">Light</option>
            <option value="dark">Dark</option>
            <option value="auto">Auto (System)</option>
          </select>
        </div>
        <div className="setting-item">
          <label>Language</label>
          <select
            value={settings.language}
            onChange={(e) => updateSetting('language', e.target.value)}
          >
            <option value="en">English</option>
            <option value="es">Español</option>
            <option value="fr">Français</option>
            <option value="de">Deutsch</option>
          </select>
        </div>
      </div>

      <div className="settings-section">
        <h3>Application Behavior</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.autoStart}
              onChange={(e) => updateSetting('autoStart', e.target.checked)}
            />
            Start Credence automatically on system boot
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.minimizeToTray}
              onChange={(e) => updateSetting('minimizeToTray', e.target.checked)}
            />
            Minimize to system tray instead of closing
          </label>
        </div>
      </div>
    </div>
  );

  const renderSecurityTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>App Lock</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.enableAppLock}
              onChange={(e) => updateSetting('enableAppLock', e.target.checked)}
            />
            Enable app-level password/PIN protection
          </label>
        </div>
        <div className="setting-item">
          <label>Auto-lock timeout</label>
          <select
            value={settings.autoLockMinutes}
            onChange={(e) => updateSetting('autoLockMinutes', Number(e.target.value))}
          >
            <option value={1}>1 minute</option>
            <option value={5}>5 minutes</option>
            <option value={10}>10 minutes</option>
            <option value={30}>30 minutes</option>
            <option value={60}>1 hour</option>
          </select>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.lockOnMinimize}
              onChange={(e) => updateSetting('lockOnMinimize', e.target.checked)}
            />
            Lock wallet when minimized
          </label>
        </div>
      </div>

      <div className="settings-section">
        <h3>Authentication</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.enableBiometric}
              onChange={(e) => updateSetting('enableBiometric', e.target.checked)}
            />
            Enable biometric authentication (if available)
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.requireAuthForExports}
              onChange={(e) => updateSetting('requireAuthForExports', e.target.checked)}
            />
            Require authentication for data exports
          </label>
        </div>
        <div className="setting-item">
          <label>Session timeout</label>
          <select
            value={settings.sessionTimeoutMinutes}
            onChange={(e) => updateSetting('sessionTimeoutMinutes', Number(e.target.value))}
          >
            <option value={30}>30 minutes</option>
            <option value={60}>1 hour</option>
            <option value={120}>2 hours</option>
            <option value={240}>4 hours</option>
            <option value={480}>8 hours</option>
          </select>
        </div>
      </div>
    </div>
  );

  const renderNetworkTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>Wallet Service</h3>
        <div className="setting-item">
          <label>Wallet Service URL</label>
          <input
            type="text"
            value={settings.walletServiceUrl}
            onChange={(e) => updateSetting('walletServiceUrl', e.target.value)}
            placeholder="http://localhost:8080"
          />
        </div>
        <div className="setting-item">
          <label>Network timeout (seconds)</label>
          <input
            type="number"
            value={settings.networkTimeout}
            onChange={(e) => updateSetting('networkTimeout', Number(e.target.value))}
            min={5}
            max={300}
          />
        </div>
      </div>

      <div className="settings-section">
        <h3>P2P Network</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.autoConnectPeers}
              onChange={(e) => updateSetting('autoConnectPeers', e.target.checked)}
            />
            Automatically connect to peers
          </label>
        </div>
        <div className="setting-item">
          <label>Maximum peers</label>
          <input
            type="number"
            value={settings.maxPeers}
            onChange={(e) => updateSetting('maxPeers', Number(e.target.value))}
            min={1}
            max={100}
          />
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.enableP2PLogging}
              onChange={(e) => updateSetting('enableP2PLogging', e.target.checked)}
            />
            Enable P2P debug logging
          </label>
        </div>
      </div>
    </div>
  );

  const renderPrivacyTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>Data Collection</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.shareAnalytics}
              onChange={(e) => updateSetting('shareAnalytics', e.target.checked)}
            />
            Share anonymous usage analytics
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.enableTelemetry}
              onChange={(e) => updateSetting('enableTelemetry', e.target.checked)}
            />
            Enable telemetry for error reporting
          </label>
        </div>
      </div>

      <div className="settings-section">
        <h3>Data Storage</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.storeEventHistory}
              onChange={(e) => updateSetting('storeEventHistory', e.target.checked)}
            />
            Store complete event history locally
          </label>
        </div>
        <div className="setting-item">
          <label>Automatic backup frequency</label>
          <select
            value={settings.autoBackupInterval}
            onChange={(e) => updateSetting('autoBackupInterval', e.target.value as any)}
          >
            <option value="never">Never</option>
            <option value="daily">Daily</option>
            <option value="weekly">Weekly</option>
            <option value="monthly">Monthly</option>
          </select>
        </div>
      </div>
    </div>
  );

  const renderNotificationsTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>General Notifications</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.enableNotifications}
              onChange={(e) => updateSetting('enableNotifications', e.target.checked)}
            />
            Enable desktop notifications
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.soundEnabled}
              onChange={(e) => updateSetting('soundEnabled', e.target.checked)}
            />
            Play notification sounds
          </label>
        </div>
      </div>

      <div className="settings-section">
        <h3>Notification Types</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.securityAlerts}
              onChange={(e) => updateSetting('securityAlerts', e.target.checked)}
            />
            Security alerts and warnings
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.eventNotifications}
              onChange={(e) => updateSetting('eventNotifications', e.target.checked)}
            />
            Event notifications (vouches, reports)
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.networkUpdates}
              onChange={(e) => updateSetting('networkUpdates', e.target.checked)}
            />
            Network status updates
          </label>
        </div>
      </div>
    </div>
  );

  const renderBackupTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>Current Data</h3>
        <div className="backup-stats">
          <div className="backup-stat">
            <span className="stat-number">{keys.length}</span>
            <span className="stat-label">Keys</span>
          </div>
          <div className="backup-stat">
            <span className="stat-number">{dids.length}</span>
            <span className="stat-label">DIDs</span>
          </div>
          <div className="backup-stat">
            <span className="stat-number">{credentials.length}</span>
            <span className="stat-label">Credentials</span>
          </div>
          <div className="backup-stat">
            <span className="stat-number">{events.length}</span>
            <span className="stat-label">Events</span>
          </div>
        </div>
      </div>

      <div className="settings-section">
        <h3>Backup Operations</h3>
        <div className="setting-item">
          <button
            className="primary-button"
            onClick={handleFullBackup}
            disabled={isBackingUp || (keys.length + dids.length + credentials.length + events.length) === 0}
          >
            {isBackingUp ? 'Creating Backup...' : '📦 Create Full Backup'}
          </button>
          <div className="setting-help">
            Export all wallet data including keys, DIDs, credentials, events, and settings
          </div>
        </div>
        <div className="setting-item">
          <input
            type="file"
            accept=".json"
            onChange={handleRestoreFile}
            style={{ display: 'none' }}
            id="restore-file-input"
            disabled={isRestoring}
          />
          <button
            className="secondary-button"
            onClick={() => document.getElementById('restore-file-input')?.click()}
            disabled={isRestoring}
          >
            {isRestoring ? 'Restoring...' : '📂 Restore from Backup'}
          </button>
          <div className="setting-help">
            Import wallet data from a previously created backup file
          </div>
        </div>
      </div>
    </div>
  );

  const renderAdvancedTab = () => (
    <div className="settings-tab-content">
      <div className="settings-section">
        <h3>Developer Options</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.debugMode}
              onChange={(e) => updateSetting('debugMode', e.target.checked)}
            />
            Enable debug mode
          </label>
        </div>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.developerTools}
              onChange={(e) => updateSetting('developerTools', e.target.checked)}
            />
            Enable developer tools
          </label>
        </div>
      </div>

      <div className="settings-section">
        <h3>Experimental Features</h3>
        <div className="setting-item checkbox">
          <label>
            <input
              type="checkbox"
              checked={settings.experimentalFeatures}
              onChange={(e) => updateSetting('experimentalFeatures', e.target.checked)}
            />
            Enable experimental features (may be unstable)
          </label>
        </div>
        <div className="setting-item">
          <label>Custom RPC endpoint</label>
          <input
            type="text"
            value={settings.customRpcEndpoint}
            onChange={(e) => updateSetting('customRpcEndpoint', e.target.value)}
            placeholder="https://custom-endpoint.com"
          />
        </div>
      </div>

      <div className="settings-section danger">
        <h3>Reset Options</h3>
        <div className="setting-item">
          <button className="danger-button" onClick={resetSettings}>
            Reset All Settings
          </button>
          <div className="setting-help">
            Reset all settings to their default values. This cannot be undone.
          </div>
        </div>
      </div>
    </div>
  );

  const renderTabContent = () => {
    switch (activeTab) {
      case 'general': return renderGeneralTab();
      case 'security': return renderSecurityTab();
      case 'network': return renderNetworkTab();
      case 'privacy': return renderPrivacyTab();
      case 'notifications': return renderNotificationsTab();
      case 'backup': return renderBackupTab();
      case 'advanced': return renderAdvancedTab();
      default: return renderGeneralTab();
    }
  };

  return (
    <div className="page settings">
      <div className="page-header">
        <h1 className="page-title">Settings</h1>
        <p className="page-subtitle">Configure your wallet preferences and behavior</p>
        {hasChanges && (
          <div className="settings-actions">
            <button className="save-button primary" onClick={saveSettings}>
              Save Changes
            </button>
            <button 
              className="discard-button secondary" 
              onClick={() => window.location.reload()}
            >
              Discard Changes
            </button>
          </div>
        )}
      </div>

      <div className="page-content">
        <div className="settings-layout">
          <div className="settings-sidebar">
            <div className="settings-tabs">
              {tabs.map(tab => (
                <button
                  key={tab.id}
                  className={`settings-tab ${activeTab === tab.id ? 'active' : ''}`}
                  onClick={() => setActiveTab(tab.id)}
                >
                  <span className="tab-icon">{tab.icon}</span>
                  <span className="tab-label">{tab.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="settings-main">
            {renderTabContent()}
          </div>
        </div>
      </div>

      <div className="settings-footer">
        <div className="version-info">
          <span>Credence Wallet v1.0.0</span>
          <span>•</span>
          <a href="#" onClick={(e) => { e.preventDefault(); /* Open about */ }}>
            About
          </a>
          <span>•</span>
          <a href="#" onClick={(e) => { e.preventDefault(); /* Open help */ }}>
            Help & Support
          </a>
        </div>
      </div>
    </div>
  );
}