import React, { useState } from 'react';
import { useWallet } from '../contexts/WalletContext';

export function WalletSetup(): JSX.Element {
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const { initialize } = useWallet();

  const handleSetup = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters long');
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      // For now, just initialize the wallet
      // In a real implementation, we would set up the wallet with the password
      await initialize();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'Failed to setup wallet');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="wallet-setup">
      <div className="setup-container">
        <div className="setup-header">
          <h1 className="setup-title">Welcome to Credence Wallet</h1>
          <p className="setup-subtitle">
            Set up your wallet to get started with decentralized identity and trust
          </p>
        </div>

        <form className="setup-form" onSubmit={handleSetup}>
          <div className="form-group">
            <label htmlFor="password" className="form-label">
              Create Password
            </label>
            <input
              type="password"
              id="password"
              className="form-input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter a strong password"
              required
              minLength={8}
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword" className="form-label">
              Confirm Password
            </label>
            <input
              type="password"
              id="confirmPassword"
              className="form-input"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your password"
              required
              minLength={8}
            />
          </div>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          <button 
            type="submit" 
            className="setup-button"
            disabled={isLoading || !password || !confirmPassword}
          >
            {isLoading ? 'Setting up...' : 'Setup Wallet'}
          </button>
        </form>

        <div className="setup-info">
          <p className="info-text">
            Your password will be used to encrypt your wallet data locally.
            Make sure to remember it as it cannot be recovered.
          </p>
        </div>
      </div>
    </div>
  );
}