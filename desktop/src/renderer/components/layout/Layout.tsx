import React, { ReactNode } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { NotificationCenter } from '../notifications/NotificationCenter';
import { useWallet } from '../../contexts/WalletContext';

interface LayoutProps {
  children: ReactNode;
}

export function Layout({ children }: LayoutProps): JSX.Element {
  const { networkStatus, isLocked } = useWallet();

  return (
    <div className="layout">
      <Header 
        isLocked={isLocked}
        networkConnected={networkStatus.walletService}
      />
      
      <div className="layout-body">
        <Sidebar />
        
        <main className="main-content">
          <div className="content-wrapper">
            {children}
          </div>
        </main>
      </div>
      
      <NotificationCenter />
    </div>
  );
}