import React from 'react';
import { NavLink } from 'react-router-dom';

const navigationItems = [
  {
    path: '/dashboard',
    label: 'Dashboard',
    icon: '📊',
  },
  {
    path: '/keys',
    label: 'Keys',
    icon: '🔑',
  },
  {
    path: '/dids',
    label: 'DIDs',
    icon: '🆔',
  },
  {
    path: '/credentials',
    label: 'Credentials',
    icon: '📜',
  },
  {
    path: '/events',
    label: 'Events',
    icon: '📝',
  },
  {
    path: '/trust-scores',
    label: 'Trust Scores',
    icon: '⭐',
  },
  {
    path: '/network',
    label: 'Network',
    icon: '🌐',
  },
  {
    path: '/settings',
    label: 'Settings',
    icon: '⚙️',
  },
];

export function Sidebar(): JSX.Element {
  return (
    <aside className="sidebar">
      <nav className="sidebar-nav">
        <ul className="nav-list">
          {navigationItems.map((item) => (
            <li key={item.path} className="nav-item">
              <NavLink
                to={item.path}
                className={({ isActive }) =>
                  `nav-link ${isActive ? 'active' : ''}`
                }
              >
                <span className="nav-icon">{item.icon}</span>
                <span className="nav-label">{item.label}</span>
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>

      <div className="sidebar-footer">
        <div className="version-info">
          <span className="version-label">Version</span>
          <span className="version-number">1.0.0</span>
        </div>
      </div>
    </aside>
  );
}