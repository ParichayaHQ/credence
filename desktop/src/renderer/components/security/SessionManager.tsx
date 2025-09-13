import React, { createContext, useContext, useReducer, useEffect } from 'react';
import { useNotification } from '../../contexts/NotificationContext';

interface SessionState {
  isActive: boolean;
  startTime: Date | null;
  lastActivity: Date | null;
  sessionId: string | null;
  userId: string | null;
  permissions: string[];
  expiresAt: Date | null;
  warningShown: boolean;
}

interface SessionConfig {
  maxInactiveMinutes: number;
  sessionTimeoutMinutes: number;
  warningMinutes: number;
  extendOnActivity: boolean;
  requireReauth: boolean;
}

type SessionAction =
  | { type: 'START_SESSION'; payload: { userId: string; permissions: string[] } }
  | { type: 'END_SESSION' }
  | { type: 'UPDATE_ACTIVITY' }
  | { type: 'SHOW_WARNING' }
  | { type: 'EXTEND_SESSION' }
  | { type: 'EXPIRE_SESSION' };

const initialState: SessionState = {
  isActive: false,
  startTime: null,
  lastActivity: null,
  sessionId: null,
  userId: null,
  permissions: [],
  expiresAt: null,
  warningShown: false
};

const defaultConfig: SessionConfig = {
  maxInactiveMinutes: 15,
  sessionTimeoutMinutes: 120, // 2 hours
  warningMinutes: 5,
  extendOnActivity: true,
  requireReauth: false
};

function sessionReducer(state: SessionState, action: SessionAction): SessionState {
  switch (action.type) {
    case 'START_SESSION':
      const now = new Date();
      const expiresAt = new Date(now.getTime() + defaultConfig.sessionTimeoutMinutes * 60 * 1000);
      return {
        ...state,
        isActive: true,
        startTime: now,
        lastActivity: now,
        sessionId: generateSessionId(),
        userId: action.payload.userId,
        permissions: action.payload.permissions,
        expiresAt,
        warningShown: false
      };
    
    case 'END_SESSION':
      return {
        ...initialState
      };
    
    case 'UPDATE_ACTIVITY':
      const activityTime = new Date();
      let newExpiresAt = state.expiresAt;
      
      if (defaultConfig.extendOnActivity && state.expiresAt) {
        // Extend session if more than half the time has passed
        const timeUntilExpiry = state.expiresAt.getTime() - activityTime.getTime();
        const sessionDuration = defaultConfig.sessionTimeoutMinutes * 60 * 1000;
        
        if (timeUntilExpiry < sessionDuration / 2) {
          newExpiresAt = new Date(activityTime.getTime() + sessionDuration);
        }
      }
      
      return {
        ...state,
        lastActivity: activityTime,
        expiresAt: newExpiresAt,
        warningShown: false
      };
    
    case 'SHOW_WARNING':
      return {
        ...state,
        warningShown: true
      };
    
    case 'EXTEND_SESSION':
      const extendedTime = new Date();
      return {
        ...state,
        lastActivity: extendedTime,
        expiresAt: new Date(extendedTime.getTime() + defaultConfig.sessionTimeoutMinutes * 60 * 1000),
        warningShown: false
      };
    
    case 'EXPIRE_SESSION':
      return {
        ...initialState
      };
    
    default:
      return state;
  }
}

function generateSessionId(): string {
  return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

interface SessionContextType {
  session: SessionState;
  startSession: (userId: string, permissions: string[]) => void;
  endSession: () => void;
  extendSession: () => void;
  hasPermission: (permission: string) => boolean;
  getTimeUntilExpiry: () => number;
  isExpiringSoon: () => boolean;
}

const SessionContext = createContext<SessionContextType | undefined>(undefined);

interface SessionManagerProps {
  children: React.ReactNode;
  config?: Partial<SessionConfig>;
  onSessionExpired?: () => void;
  onSessionWarning?: (minutesLeft: number) => void;
}

export function SessionManager({ 
  children, 
  config = {},
  onSessionExpired,
  onSessionWarning
}: SessionManagerProps) {
  const [session, dispatch] = useReducer(sessionReducer, initialState);
  const { showNotification } = useNotification();

  const finalConfig = { ...defaultConfig, ...config };

  // Load session from localStorage on mount
  useEffect(() => {
    const savedSession = localStorage.getItem('walletSession');
    if (savedSession) {
      try {
        const parsed = JSON.parse(savedSession);
        const expiresAt = new Date(parsed.expiresAt);
        
        if (expiresAt > new Date()) {
          dispatch({
            type: 'START_SESSION',
            payload: {
              userId: parsed.userId,
              permissions: parsed.permissions
            }
          });
        } else {
          localStorage.removeItem('walletSession');
        }
      } catch (error) {
        console.error('Failed to restore session:', error);
        localStorage.removeItem('walletSession');
      }
    }
  }, []);

  // Save session to localStorage
  useEffect(() => {
    if (session.isActive) {
      const sessionData = {
        userId: session.userId,
        permissions: session.permissions,
        expiresAt: session.expiresAt?.toISOString(),
        sessionId: session.sessionId
      };
      localStorage.setItem('walletSession', JSON.stringify(sessionData));
    } else {
      localStorage.removeItem('walletSession');
    }
  }, [session]);

  // Session monitoring
  useEffect(() => {
    if (!session.isActive) return;

    const checkSession = () => {
      const now = new Date();
      
      if (!session.lastActivity || !session.expiresAt) return;

      // Check for inactivity timeout
      const inactiveTime = now.getTime() - session.lastActivity.getTime();
      const maxInactiveMs = finalConfig.maxInactiveMinutes * 60 * 1000;
      
      if (inactiveTime >= maxInactiveMs) {
        dispatch({ type: 'EXPIRE_SESSION' });
        showNotification('Session expired due to inactivity', 'warning');
        onSessionExpired?.();
        return;
      }

      // Check for session timeout
      if (now >= session.expiresAt) {
        dispatch({ type: 'EXPIRE_SESSION' });
        showNotification('Session expired', 'warning');
        onSessionExpired?.();
        return;
      }

      // Check for warning threshold
      const timeUntilExpiry = session.expiresAt.getTime() - now.getTime();
      const warningMs = finalConfig.warningMinutes * 60 * 1000;
      
      if (timeUntilExpiry <= warningMs && !session.warningShown) {
        const minutesLeft = Math.ceil(timeUntilExpiry / (1000 * 60));
        dispatch({ type: 'SHOW_WARNING' });
        onSessionWarning?.(minutesLeft);
        showNotification(
          `Session expires in ${minutesLeft} minutes`,
          'warning',
          'Click to extend session'
        );
      }
    };

    const interval = setInterval(checkSession, 10000); // Check every 10 seconds
    return () => clearInterval(interval);
  }, [session, finalConfig, onSessionExpired, onSessionWarning, showNotification]);

  // Track user activity
  useEffect(() => {
    if (!session.isActive) return;

    const handleActivity = () => {
      dispatch({ type: 'UPDATE_ACTIVITY' });
    };

    const events = ['mousedown', 'keydown', 'scroll', 'click', 'touchstart'];
    events.forEach(event => {
      document.addEventListener(event, handleActivity, { passive: true });
    });

    return () => {
      events.forEach(event => {
        document.removeEventListener(event, handleActivity);
      });
    };
  }, [session.isActive]);

  const startSession = (userId: string, permissions: string[] = []) => {
    dispatch({
      type: 'START_SESSION',
      payload: { userId, permissions }
    });
    
    showNotification('Secure session started', 'success');
  };

  const endSession = () => {
    dispatch({ type: 'END_SESSION' });
    showNotification('Session ended', 'info');
  };

  const extendSession = () => {
    dispatch({ type: 'EXTEND_SESSION' });
    showNotification('Session extended', 'success');
  };

  const hasPermission = (permission: string): boolean => {
    return session.permissions.includes(permission) || session.permissions.includes('*');
  };

  const getTimeUntilExpiry = (): number => {
    if (!session.expiresAt) return 0;
    return Math.max(0, session.expiresAt.getTime() - new Date().getTime());
  };

  const isExpiringSoon = (): boolean => {
    const timeLeft = getTimeUntilExpiry();
    return timeLeft > 0 && timeLeft <= finalConfig.warningMinutes * 60 * 1000;
  };

  const contextValue: SessionContextType = {
    session,
    startSession,
    endSession,
    extendSession,
    hasPermission,
    getTimeUntilExpiry,
    isExpiringSoon
  };

  return (
    <SessionContext.Provider value={contextValue}>
      {children}
    </SessionContext.Provider>
  );
}

export function useSession(): SessionContextType {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error('useSession must be used within a SessionManager');
  }
  return context;
}

// Session status component
interface SessionStatusProps {
  showExtendButton?: boolean;
  compact?: boolean;
}

export function SessionStatus({ showExtendButton = true, compact = false }: SessionStatusProps) {
  const { session, extendSession, getTimeUntilExpiry, isExpiringSoon } = useSession();

  if (!session.isActive) {
    return compact ? null : (
      <div className="session-status inactive">
        <span className="status-text">No active session</span>
      </div>
    );
  }

  const timeLeft = getTimeUntilExpiry();
  const hours = Math.floor(timeLeft / (1000 * 60 * 60));
  const minutes = Math.floor((timeLeft % (1000 * 60 * 60)) / (1000 * 60));
  
  const formatTimeLeft = () => {
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const getStatusClass = () => {
    if (isExpiringSoon()) return 'expiring';
    return 'active';
  };

  if (compact) {
    return (
      <div className={`session-status compact ${getStatusClass()}`}>
        <span className="time-left">{formatTimeLeft()}</span>
        {isExpiringSoon() && showExtendButton && (
          <button onClick={extendSession} className="extend-button">
            +
          </button>
        )}
      </div>
    );
  }

  return (
    <div className={`session-status ${getStatusClass()}`}>
      <div className="session-info">
        <span className="session-label">Session:</span>
        <span className="time-left">{formatTimeLeft()} left</span>
      </div>
      
      <div className="session-details">
        <span className="user-id">{session.userId}</span>
        <span className="session-id" title={session.sessionId || undefined}>
          {session.sessionId?.substring(0, 8) || 'N/A'}...
        </span>
      </div>
      
      {showExtendButton && (
        <button onClick={extendSession} className="extend-session-button">
          Extend
        </button>
      )}
    </div>
  );
}

export default SessionManager;