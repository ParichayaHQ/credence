import React from 'react';

export function TestApp(): JSX.Element {
  const hasWalletAPI = typeof window !== 'undefined' && 'walletAPI' in window;
  const hasSystemAPI = typeof window !== 'undefined' && 'systemAPI' in window;
  const hasEventAPI = typeof window !== 'undefined' && 'eventAPI' in window;

  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1 style={{ color: 'green' }}>🎉 Credence Wallet Debug</h1>
      <div style={{ background: '#f0f0f0', padding: '15px', borderRadius: '8px' }}>
        <h2>API Availability:</h2>
        <ul>
          <li>walletAPI: {hasWalletAPI ? '✅ Available' : '❌ Missing'}</li>
          <li>systemAPI: {hasSystemAPI ? '✅ Available' : '❌ Missing'}</li>
          <li>eventAPI: {hasEventAPI ? '✅ Available' : '❌ Missing'}</li>
        </ul>
      </div>
      
      <div style={{ marginTop: '20px', background: '#e8f4fd', padding: '15px', borderRadius: '8px' }}>
        <h2>React Status:</h2>
        <p>✅ React is rendering successfully!</p>
        <p>✅ Electron window is loading the renderer!</p>
        <p>✅ HTML and JavaScript are working!</p>
      </div>

      {hasWalletAPI && (
        <div style={{ marginTop: '20px', background: '#e8ffe8', padding: '15px', borderRadius: '8px' }}>
          <h2>WalletAPI Methods:</h2>
          <pre style={{ fontSize: '12px', overflow: 'auto' }}>
            {JSON.stringify(Object.keys((window as any).walletAPI || {}), null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}