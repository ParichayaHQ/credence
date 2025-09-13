import React from 'react';
import { useNotification } from '../../contexts/NotificationContext';
import { NotificationItem } from './NotificationItem';

export function NotificationCenter(): JSX.Element {
  const { notifications, clearAllNotifications } = useNotification();

  if (notifications.length === 0) {
    return <></>;
  }

  return (
    <div className="notification-center">
      <div className="notification-header">
        <span className="notification-title">
          Notifications ({notifications.length})
        </span>
        {notifications.length > 0 && (
          <button 
            className="clear-all-button"
            onClick={clearAllNotifications}
          >
            Clear All
          </button>
        )}
      </div>
      
      <div className="notification-list">
        {notifications.map((notification) => (
          <NotificationItem 
            key={notification.id} 
            notification={notification}
          />
        ))}
      </div>
    </div>
  );
}