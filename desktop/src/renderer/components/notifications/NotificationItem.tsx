import React from 'react';
import { Notification } from '../../contexts/NotificationContext';
import { useNotification } from '../../contexts/NotificationContext';

interface NotificationItemProps {
  notification: Notification;
}

export function NotificationItem({ notification }: NotificationItemProps): JSX.Element {
  const { removeNotification } = useNotification();

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      case 'warning':
        return '⚠️';
      case 'info':
      default:
        return 'ℹ️';
    }
  };

  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    
    if (minutes < 1) {
      return 'just now';
    } else if (minutes < 60) {
      return `${minutes}m ago`;
    } else if (hours < 24) {
      return `${hours}h ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  const handleClose = () => {
    removeNotification(notification.id);
  };

  return (
    <div className={`notification-item notification-${notification.type}`}>
      <div className="notification-content">
        <div className="notification-icon">
          {getNotificationIcon(notification.type)}
        </div>
        
        <div className="notification-body">
          <div className="notification-title">
            {notification.title}
          </div>
          
          {notification.message && (
            <div className="notification-message">
              {notification.message}
            </div>
          )}
          
          <div className="notification-timestamp">
            {formatTimestamp(notification.timestamp)}
          </div>
        </div>
      </div>
      
      <button 
        className="notification-close"
        onClick={handleClose}
        title="Dismiss notification"
      >
        ×
      </button>
    </div>
  );
}