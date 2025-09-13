import React, { useState, useEffect } from 'react';
import { useNotification } from '../../contexts/NotificationContext';

interface AppLockConfig {
  enabled: boolean;
  lockType: 'password' | 'pin';
  autoLockMinutes: number;
  requireOnStartup: boolean;
  maxFailedAttempts: number;
  lockoutMinutes: number;
}

interface LockState {
  isLocked: boolean;
  isSetup: boolean;
  failedAttempts: number;
  lockoutUntil?: Date;
  lastActivity: Date;
}

interface AppLockManagerProps {
  children: React.ReactNode;
  onLockStateChange?: (locked: boolean) => void;
}

export function AppLockManager({ children, onLockStateChange }: AppLockManagerProps) {
  const [config, setConfig] = useState<AppLockConfig>({
    enabled: false,
    lockType: 'password',
    autoLockMinutes: 5,
    requireOnStartup: true,
    maxFailedAttempts: 3,
    lockoutMinutes: 15
  });

  const [lockState, setLockState] = useState<LockState>({
    isLocked: false,
    isSetup: false,
    failedAttempts: 0,
    lastActivity: new Date()
  });

  const [currentInput, setCurrentInput] = useState('');
  const [showSetup, setShowSetup] = useState(false);
  const [setupStep, setSetupStep] = useState<'choose-type' | 'set-credentials' | 'confirm-credentials'>('choose-type');
  const [setupCredentials, setSetupCredentials] = useState('');
  const [confirmCredentials, setConfirmCredentials] = useState('');

  const { showNotification } = useNotification();

  // Load configuration and state from localStorage
  useEffect(() => {
    const savedConfig = localStorage.getItem('appLockConfig');
    const savedState = localStorage.getItem('appLockState');
    
    if (savedConfig) {
      setConfig(JSON.parse(savedConfig));
    }
    
    if (savedState) {
      const state = JSON.parse(savedState);
      setLockState({
        ...state,
        lastActivity: new Date(state.lastActivity),
        lockoutUntil: state.lockoutUntil ? new Date(state.lockoutUntil) : undefined
      });
    }

    // Check if app should be locked on startup
    const savedIsSetup = localStorage.getItem('appLockSetup') === 'true';
    if (savedIsSetup) {
      setLockState(prev => ({ ...prev, isSetup: true }));
      
      if (JSON.parse(savedConfig || '{"enabled": false, "requireOnStartup": true}').enabled &&
          JSON.parse(savedConfig || '{"requireOnStartup": true}').requireOnStartup) {
        setLockState(prev => ({ ...prev, isLocked: true }));
      }
    }
  }, []);

  // Save configuration when it changes
  useEffect(() => {
    localStorage.setItem('appLockConfig', JSON.stringify(config));
  }, [config]);

  // Save lock state when it changes
  useEffect(() => {
    localStorage.setItem('appLockState', JSON.stringify(lockState));
    onLockStateChange?.(lockState.isLocked);
  }, [lockState, onLockStateChange]);

  // Auto-lock timer
  useEffect(() => {
    if (!config.enabled || lockState.isLocked) return;

    const interval = setInterval(() => {
      const now = new Date();
      const timeSinceLastActivity = now.getTime() - lockState.lastActivity.getTime();
      const autoLockMs = config.autoLockMinutes * 60 * 1000;

      if (timeSinceLastActivity >= autoLockMs) {
        handleAutoLock();
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [config, lockState]);

  // Update last activity on user interaction
  useEffect(() => {
    const updateActivity = () => {
      if (!lockState.isLocked) {
        setLockState(prev => ({ ...prev, lastActivity: new Date() }));
      }
    };

    const events = ['mousedown', 'keydown', 'scroll', 'click'];
    events.forEach(event => {
      document.addEventListener(event, updateActivity);
    });

    return () => {
      events.forEach(event => {
        document.removeEventListener(event, updateActivity);
      });
    };
  }, [lockState.isLocked]);

  const handleAutoLock = () => {
    setLockState(prev => ({ ...prev, isLocked: true }));
    showNotification('App locked due to inactivity', 'info');
  };

  const hashCredentials = async (credentials: string): Promise<string> => {
    // In a real implementation, use a proper crypto library like bcrypt
    // This is a simplified version for demo purposes
    const encoder = new TextEncoder();
    const data = encoder.encode(credentials);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  };

  const verifyCredentials = async (input: string): Promise<boolean> => {
    const stored = localStorage.getItem('appLockCredentials');
    if (!stored) return false;

    const inputHash = await hashCredentials(input);
    return inputHash === stored;
  };

  const handleSetupComplete = async () => {
    if (setupCredentials !== confirmCredentials) {
      showNotification('Credentials do not match', 'error');
      return;
    }

    if (config.lockType === 'pin' && setupCredentials.length < 4) {
      showNotification('PIN must be at least 4 digits', 'error');
      return;
    }

    if (config.lockType === 'password' && setupCredentials.length < 6) {
      showNotification('Password must be at least 6 characters', 'error');
      return;
    }

    const hashedCredentials = await hashCredentials(setupCredentials);
    localStorage.setItem('appLockCredentials', hashedCredentials);
    localStorage.setItem('appLockSetup', 'true');

    setLockState(prev => ({ ...prev, isSetup: true, isLocked: true }));
    setShowSetup(false);
    setSetupCredentials('');
    setConfirmCredentials('');
    setSetupStep('choose-type');

    showNotification(`App lock ${config.lockType} set successfully`, 'success');
  };

  const handleUnlock = async () => {
    if (lockState.lockoutUntil && new Date() < lockState.lockoutUntil) {
      const remainingMinutes = Math.ceil((lockState.lockoutUntil.getTime() - new Date().getTime()) / (1000 * 60));
      showNotification(`Account locked. Try again in ${remainingMinutes} minutes`, 'error');
      return;
    }

    const isValid = await verifyCredentials(currentInput);

    if (isValid) {
      setLockState(prev => ({
        ...prev,
        isLocked: false,
        failedAttempts: 0,
        lockoutUntil: undefined,
        lastActivity: new Date()
      }));
      setCurrentInput('');
      showNotification('App unlocked successfully', 'success');
    } else {
      const newFailedAttempts = lockState.failedAttempts + 1;
      
      if (newFailedAttempts >= config.maxFailedAttempts) {
        const lockoutUntil = new Date(Date.now() + config.lockoutMinutes * 60 * 1000);
        setLockState(prev => ({
          ...prev,
          failedAttempts: newFailedAttempts,
          lockoutUntil
        }));
        showNotification(`Too many failed attempts. Locked for ${config.lockoutMinutes} minutes`, 'error');
      } else {
        setLockState(prev => ({ ...prev, failedAttempts: newFailedAttempts }));
        const remainingAttempts = config.maxFailedAttempts - newFailedAttempts;
        showNotification(`Invalid ${config.lockType}. ${remainingAttempts} attempts remaining`, 'error');
      }
      
      setCurrentInput('');
    }
  };

  const handleLock = () => {
    setLockState(prev => ({ ...prev, isLocked: true }));
    showNotification('App locked', 'info');
  };

  const handleDisable = () => {
    localStorage.removeItem('appLockCredentials');
    localStorage.removeItem('appLockSetup');
    setConfig(prev => ({ ...prev, enabled: false }));
    setLockState({
      isLocked: false,
      isSetup: false,
      failedAttempts: 0,
      lastActivity: new Date()
    });
    showNotification('App lock disabled', 'info');
  };

  const renderSetupWizard = () => (
    <div className="app-lock-setup">
      <div className="setup-container">
        <div className="setup-header">
          <h2>Set Up App Lock</h2>
          <p>Protect your wallet with a password or PIN</p>
        </div>

        {setupStep === 'choose-type' && (
          <div className="setup-step">
            <h3>Choose Lock Type</h3>
            <div className="lock-type-options">
              <button
                className={`lock-type-option ${config.lockType === 'password' ? 'selected' : ''}`}
                onClick={() => setConfig(prev => ({ ...prev, lockType: 'password' }))}
              >
                <div className="option-icon">🔐</div>
                <div className="option-title">Password</div>
                <div className="option-description">Use a password (6+ characters)</div>
              </button>
              
              <button
                className={`lock-type-option ${config.lockType === 'pin' ? 'selected' : ''}`}
                onClick={() => setConfig(prev => ({ ...prev, lockType: 'pin' }))}
              >
                <div className="option-icon">🔢</div>
                <div className="option-title">PIN</div>
                <div className="option-description">Use a numeric PIN (4+ digits)</div>
              </button>
            </div>

            <div className="setup-options">
              <label className="option-checkbox">
                <input
                  type="checkbox"
                  checked={config.requireOnStartup}
                  onChange={(e) => setConfig(prev => ({ ...prev, requireOnStartup: e.target.checked }))}
                />
                <span>Require unlock on app startup</span>
              </label>

              <div className="option-select">
                <label>Auto-lock after:</label>
                <select
                  value={config.autoLockMinutes}
                  onChange={(e) => setConfig(prev => ({ ...prev, autoLockMinutes: Number(e.target.value) }))}
                >
                  <option value={1}>1 minute</option>
                  <option value={5}>5 minutes</option>
                  <option value={10}>10 minutes</option>
                  <option value={15}>15 minutes</option>
                  <option value={30}>30 minutes</option>
                  <option value={60}>1 hour</option>
                </select>
              </div>
            </div>

            <div className="setup-actions">
              <button
                onClick={() => setSetupStep('set-credentials')}
                className="setup-button primary"
              >
                Continue
              </button>
              <button
                onClick={() => setShowSetup(false)}
                className="setup-button secondary"
              >
                Cancel
              </button>
            </div>
          </div>
        )}

        {setupStep === 'set-credentials' && (
          <div className="setup-step">
            <h3>Set {config.lockType === 'password' ? 'Password' : 'PIN'}</h3>
            <div className="credentials-input">
              <input
                type={config.lockType === 'password' ? 'password' : 'text'}
                value={setupCredentials}
                onChange={(e) => setSetupCredentials(e.target.value)}
                placeholder={`Enter your ${config.lockType}`}
                pattern={config.lockType === 'pin' ? '[0-9]*' : undefined}
                maxLength={config.lockType === 'pin' ? 10 : undefined}
                className="setup-input"
                autoFocus
              />
              <div className="input-hint">
                {config.lockType === 'password' 
                  ? 'Minimum 6 characters' 
                  : 'Minimum 4 digits'
                }
              </div>
            </div>

            <div className="setup-actions">
              <button
                onClick={() => setSetupStep('confirm-credentials')}
                disabled={
                  (config.lockType === 'password' && setupCredentials.length < 6) ||
                  (config.lockType === 'pin' && setupCredentials.length < 4)
                }
                className="setup-button primary"
              >
                Next
              </button>
              <button
                onClick={() => setSetupStep('choose-type')}
                className="setup-button secondary"
              >
                Back
              </button>
            </div>
          </div>
        )}

        {setupStep === 'confirm-credentials' && (
          <div className="setup-step">
            <h3>Confirm {config.lockType === 'password' ? 'Password' : 'PIN'}</h3>
            <div className="credentials-input">
              <input
                type={config.lockType === 'password' ? 'password' : 'text'}
                value={confirmCredentials}
                onChange={(e) => setConfirmCredentials(e.target.value)}
                placeholder={`Confirm your ${config.lockType}`}
                pattern={config.lockType === 'pin' ? '[0-9]*' : undefined}
                maxLength={config.lockType === 'pin' ? 10 : undefined}
                className="setup-input"
                autoFocus
              />
              <div className="input-validation">
                {confirmCredentials && (
                  <span className={setupCredentials === confirmCredentials ? 'valid' : 'invalid'}>
                    {setupCredentials === confirmCredentials ? '✅ Match' : '❌ Does not match'}
                  </span>
                )}
              </div>
            </div>

            <div className="setup-actions">
              <button
                onClick={handleSetupComplete}
                disabled={!confirmCredentials || setupCredentials !== confirmCredentials}
                className="setup-button primary"
              >
                Set Up Lock
              </button>
              <button
                onClick={() => setSetupStep('set-credentials')}
                className="setup-button secondary"
              >
                Back
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  const renderLockScreen = () => {
    const isLockedOut = lockState.lockoutUntil && new Date() < lockState.lockoutUntil;
    const lockoutMinutes = isLockedOut 
      ? Math.ceil((lockState.lockoutUntil!.getTime() - new Date().getTime()) / (1000 * 60))
      : 0;

    return (
      <div className="app-lock-screen">
        <div className="lock-container">
          <div className="lock-header">
            <div className="lock-icon">🔒</div>
            <h2>Credence Wallet</h2>
            <p>Enter your {config.lockType} to unlock</p>
          </div>

          {!isLockedOut ? (
            <div className="unlock-form">
              <div className="credentials-input">
                <input
                  type={config.lockType === 'password' ? 'password' : 'text'}
                  value={currentInput}
                  onChange={(e) => setCurrentInput(e.target.value)}
                  placeholder={`Enter ${config.lockType}`}
                  pattern={config.lockType === 'pin' ? '[0-9]*' : undefined}
                  maxLength={config.lockType === 'pin' ? 10 : undefined}
                  className="unlock-input"
                  autoFocus
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleUnlock();
                    }
                  }}
                />
              </div>

              {lockState.failedAttempts > 0 && (
                <div className="failed-attempts">
                  Failed attempts: {lockState.failedAttempts} / {config.maxFailedAttempts}
                </div>
              )}

              <div className="unlock-actions">
                <button
                  onClick={handleUnlock}
                  disabled={!currentInput}
                  className="unlock-button"
                >
                  Unlock
                </button>
              </div>
            </div>
          ) : (
            <div className="lockout-message">
              <div className="lockout-icon">⏱️</div>
              <h3>Account Temporarily Locked</h3>
              <p>Too many failed attempts. Try again in {lockoutMinutes} minutes.</p>
            </div>
          )}
        </div>
      </div>
    );
  };

  // Public API for other components
  const appLockAPI = {
    isEnabled: config.enabled,
    isLocked: lockState.isLocked,
    isSetup: lockState.isSetup,
    lockType: config.lockType,
    lock: handleLock,
    showSetup: () => setShowSetup(true),
    disable: handleDisable,
    updateConfig: (newConfig: Partial<AppLockConfig>) => {
      setConfig(prev => ({ ...prev, ...newConfig }));
    }
  };

  // Expose API to window for other components to use
  useEffect(() => {
    (window as any).appLockAPI = appLockAPI;
  }, [appLockAPI]);

  if (showSetup) {
    return renderSetupWizard();
  }

  if (config.enabled && lockState.isLocked) {
    return renderLockScreen();
  }

  return <>{children}</>;
}

export default AppLockManager;