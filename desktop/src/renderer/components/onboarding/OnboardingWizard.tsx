import React, { useState, useEffect } from 'react';
import { useWallet } from '../../contexts/WalletContext';
import { useNotification } from '../../contexts/NotificationContext';

interface OnboardingStep {
  id: string;
  title: string;
  subtitle: string;
  component: React.ComponentType<OnboardingStepProps>;
  canSkip: boolean;
  isComplete: boolean;
  requirements?: string[];
}

interface OnboardingStepProps {
  onNext: () => void;
  onSkip?: () => void;
  onComplete: (data?: any) => void;
  stepData: any;
  setStepData: (data: any) => void;
}

interface OnboardingProgress {
  currentStep: number;
  completedSteps: string[];
  skippedSteps: string[];
  userData: Record<string, any>;
  isComplete: boolean;
}

// Welcome Step Component
function WelcomeStep({ onNext }: OnboardingStepProps) {
  return (
    <div className="onboarding-step welcome-step">
      <div className="welcome-content">
        <div className="welcome-icon">🎉</div>
        <h2>Welcome to Credence Wallet</h2>
        <p className="welcome-description">
          Your gateway to decentralized identity and verifiable credentials. 
          Let's get you set up in just a few steps.
        </p>
        
        <div className="features-preview">
          <div className="feature-item">
            <span className="feature-icon">🔑</span>
            <span className="feature-text">Secure key management</span>
          </div>
          <div className="feature-item">
            <span className="feature-icon">🆔</span>
            <span className="feature-text">Decentralized identity (DIDs)</span>
          </div>
          <div className="feature-item">
            <span className="feature-icon">📜</span>
            <span className="feature-text">Verifiable credentials</span>
          </div>
          <div className="feature-item">
            <span className="feature-icon">⭐</span>
            <span className="feature-text">Trust score management</span>
          </div>
        </div>

        <div className="welcome-actions">
          <button onClick={onNext} className="welcome-button primary">
            Get Started
          </button>
        </div>
      </div>
    </div>
  );
}

// Security Setup Step
function SecuritySetupStep({ onNext, onSkip, stepData, setStepData }: OnboardingStepProps) {
  const [backupConfirmed, setBackupConfirmed] = useState(false);
  
  return (
    <div className="onboarding-step security-step">
      <div className="step-content">
        <h2>Security Setup</h2>
        <p>Protect your wallet with these important security features.</p>

        <div className="security-options">
          <div className="security-option">
            <div className="option-header">
              <span className="option-icon">🔐</span>
              <h3>App Lock Protection</h3>
            </div>
            <p>Add a password or PIN to protect access to your wallet.</p>
            <div className="option-controls">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={stepData.enableAppLock || false}
                  onChange={(e) => setStepData({ ...stepData, enableAppLock: e.target.checked })}
                />
                <span>Enable app lock (recommended)</span>
              </label>
            </div>
          </div>

          <div className="security-option">
            <div className="option-header">
              <span className="option-icon">💾</span>
              <h3>Backup Reminder</h3>
            </div>
            <p>Your keys and data should be backed up regularly to prevent loss.</p>
            <div className="option-controls">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={backupConfirmed}
                  onChange={(e) => setBackupConfirmed(e.target.checked)}
                />
                <span>I understand the importance of regular backups</span>
              </label>
            </div>
          </div>

          <div className="security-option">
            <div className="option-header">
              <span className="option-icon">🔔</span>
              <h3>Security Notifications</h3>
            </div>
            <p>Get notified about important security events and activities.</p>
            <div className="option-controls">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={stepData.enableNotifications !== false}
                  onChange={(e) => setStepData({ ...stepData, enableNotifications: e.target.checked })}
                />
                <span>Enable security notifications</span>
              </label>
            </div>
          </div>
        </div>

        <div className="security-warning">
          <div className="warning-icon">⚠️</div>
          <div className="warning-content">
            <strong>Important:</strong> Keep your wallet secure by enabling app lock and 
            creating regular backups. Lost keys cannot be recovered.
          </div>
        </div>

        <div className="step-actions">
          <button
            onClick={onNext}
            disabled={!backupConfirmed}
            className="step-button primary"
          >
            Continue
          </button>
          <button onClick={onSkip} className="step-button secondary">
            Skip for Now
          </button>
        </div>
      </div>
    </div>
  );
}

// Key Generation Step
function KeyGenerationStep({ onNext, stepData, setStepData }: OnboardingStepProps) {
  const [keyType, setKeyType] = useState('Ed25519');
  const [isGenerating, setIsGenerating] = useState(false);
  const [keyGenerated, setKeyGenerated] = useState(false);
  
  const handleGenerateKey = async () => {
    setIsGenerating(true);
    try {
      // Simulate key generation
      await new Promise(resolve => setTimeout(resolve, 2000));
      setKeyGenerated(true);
      setStepData({ ...stepData, keyType, keyGenerated: true });
    } catch (error) {
      console.error('Key generation failed:', error);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="onboarding-step key-generation-step">
      <div className="step-content">
        <h2>Create Your First Key</h2>
        <p>Generate a cryptographic key pair for your decentralized identity.</p>

        {!keyGenerated ? (
          <div className="key-generation-form">
            <div className="key-type-selection">
              <label>Key Type:</label>
              <select
                value={keyType}
                onChange={(e) => setKeyType(e.target.value)}
                disabled={isGenerating}
              >
                <option value="Ed25519">Ed25519 (recommended)</option>
                <option value="secp256k1">secp256k1</option>
              </select>
            </div>

            <div className="key-info">
              <h4>What is Ed25519?</h4>
              <ul>
                <li>Fast and secure elliptic curve algorithm</li>
                <li>Small key size with strong security</li>
                <li>Widely supported for DID operations</li>
                <li>Recommended for most use cases</li>
              </ul>
            </div>

            <div className="generation-actions">
              <button
                onClick={handleGenerateKey}
                disabled={isGenerating}
                className="generate-button primary"
              >
                {isGenerating ? (
                  <>
                    <span className="loading-spinner">⟳</span>
                    Generating Key...
                  </>
                ) : (
                  'Generate Key Pair'
                )}
              </button>
            </div>
          </div>
        ) : (
          <div className="key-success">
            <div className="success-icon">✅</div>
            <h3>Key Generated Successfully!</h3>
            <p>Your {keyType} key pair has been created and securely stored.</p>
            
            <div className="key-details">
              <div className="detail-item">
                <span className="label">Key Type:</span>
                <span className="value">{keyType}</span>
              </div>
              <div className="detail-item">
                <span className="label">Use Case:</span>
                <span className="value">DID operations, credential signing</span>
              </div>
            </div>

            <div className="security-reminder">
              <span className="reminder-icon">🔒</span>
              <span>Your private key is encrypted and stored securely on this device.</span>
            </div>
          </div>
        )}

        <div className="step-actions">
          <button
            onClick={onNext}
            disabled={!keyGenerated}
            className="step-button primary"
          >
            Continue
          </button>
        </div>
      </div>
    </div>
  );
}

// DID Creation Step
function DIDCreationStep({ onNext, stepData, setStepData }: OnboardingStepProps) {
  const [didMethod, setDidMethod] = useState('did:key');
  const [isCreating, setIsCreating] = useState(false);
  const [didCreated, setDidCreated] = useState(false);
  const [createdDID, setCreatedDID] = useState('');

  const handleCreateDID = async () => {
    setIsCreating(true);
    try {
      // Simulate DID creation
      await new Promise(resolve => setTimeout(resolve, 1500));
      const mockDID = `did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK`;
      setCreatedDID(mockDID);
      setDidCreated(true);
      setStepData({ ...stepData, didMethod, createdDID: mockDID });
    } catch (error) {
      console.error('DID creation failed:', error);
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="onboarding-step did-creation-step">
      <div className="step-content">
        <h2>Create Your Decentralized Identity</h2>
        <p>Generate a DID (Decentralized Identifier) for your digital identity.</p>

        {!didCreated ? (
          <div className="did-creation-form">
            <div className="did-method-selection">
              <label>DID Method:</label>
              <select
                value={didMethod}
                onChange={(e) => setDidMethod(e.target.value)}
                disabled={isCreating}
              >
                <option value="did:key">did:key (recommended)</option>
                <option value="did:web">did:web (future)</option>
              </select>
            </div>

            <div className="did-info">
              <h4>About did:key</h4>
              <ul>
                <li>No blockchain or registry required</li>
                <li>DID derived directly from your public key</li>
                <li>Instant creation and resolution</li>
                <li>Perfect for getting started</li>
              </ul>
            </div>

            <div className="creation-actions">
              <button
                onClick={handleCreateDID}
                disabled={isCreating || !stepData.keyGenerated}
                className="create-button primary"
              >
                {isCreating ? (
                  <>
                    <span className="loading-spinner">⟳</span>
                    Creating DID...
                  </>
                ) : (
                  'Create DID'
                )}
              </button>
            </div>
          </div>
        ) : (
          <div className="did-success">
            <div className="success-icon">🆔</div>
            <h3>DID Created Successfully!</h3>
            <p>Your decentralized identifier is ready to use.</p>
            
            <div className="did-display">
              <label>Your DID:</label>
              <div className="did-value">
                <code>{createdDID}</code>
                <button
                  onClick={() => navigator.clipboard.writeText(createdDID)}
                  className="copy-button"
                  title="Copy to clipboard"
                >
                  📋
                </button>
              </div>
            </div>

            <div className="did-capabilities">
              <h4>What you can do with your DID:</h4>
              <ul>
                <li>Receive and store verifiable credentials</li>
                <li>Create presentations for verification</li>
                <li>Participate in trust score calculations</li>
                <li>Vouch for or report other identities</li>
              </ul>
            </div>
          </div>
        )}

        <div className="step-actions">
          <button
            onClick={onNext}
            disabled={!didCreated}
            className="step-button primary"
          >
            Continue
          </button>
        </div>
      </div>
    </div>
  );
}

// Completion Step
function CompletionStep({ onComplete, stepData }: OnboardingStepProps) {
  const handleFinish = () => {
    onComplete(stepData);
  };

  return (
    <div className="onboarding-step completion-step">
      <div className="step-content">
        <div className="completion-icon">🎊</div>
        <h2>Welcome to Credence!</h2>
        <p>Your wallet is now set up and ready to use.</p>

        <div className="setup-summary">
          <h3>What we've accomplished:</h3>
          <div className="summary-items">
            <div className="summary-item completed">
              <span className="item-icon">✅</span>
              <span className="item-text">Generated your first cryptographic key</span>
            </div>
            <div className="summary-item completed">
              <span className="item-icon">✅</span>
              <span className="item-text">Created your decentralized identity (DID)</span>
            </div>
            {stepData.enableAppLock && (
              <div className="summary-item completed">
                <span className="item-icon">✅</span>
                <span className="item-text">Configured app lock security</span>
              </div>
            )}
            <div className="summary-item completed">
              <span className="item-icon">✅</span>
              <span className="item-text">Set up secure storage</span>
            </div>
          </div>
        </div>

        <div className="next-steps">
          <h3>Next steps you might want to explore:</h3>
          <ul>
            <li>Import your first verifiable credential</li>
            <li>Connect with other identities in your network</li>
            <li>Explore trust score features</li>
            <li>Set up backup and recovery options</li>
          </ul>
        </div>

        <div className="completion-actions">
          <button onClick={handleFinish} className="finish-button primary">
            Start Using Credence
          </button>
        </div>
      </div>
    </div>
  );
}

// Main Onboarding Wizard Component
export function OnboardingWizard() {
  const [progress, setProgress] = useState<OnboardingProgress>({
    currentStep: 0,
    completedSteps: [],
    skippedSteps: [],
    userData: {},
    isComplete: false
  });

  const { showNotification } = useNotification();

  const steps: OnboardingStep[] = [
    {
      id: 'welcome',
      title: 'Welcome',
      subtitle: 'Introduction to Credence Wallet',
      component: WelcomeStep,
      canSkip: false,
      isComplete: false
    },
    {
      id: 'security',
      title: 'Security Setup',
      subtitle: 'Configure wallet protection',
      component: SecuritySetupStep,
      canSkip: true,
      isComplete: false
    },
    {
      id: 'key-generation',
      title: 'Key Generation',
      subtitle: 'Create your first key pair',
      component: KeyGenerationStep,
      canSkip: false,
      isComplete: false,
      requirements: ['security']
    },
    {
      id: 'did-creation',
      title: 'DID Creation',
      subtitle: 'Generate your decentralized identity',
      component: DIDCreationStep,
      canSkip: false,
      isComplete: false,
      requirements: ['key-generation']
    },
    {
      id: 'completion',
      title: 'Complete',
      subtitle: 'Finish setup and start using Credence',
      component: CompletionStep,
      canSkip: false,
      isComplete: false,
      requirements: ['did-creation']
    }
  ];

  useEffect(() => {
    // Load saved progress
    const saved = localStorage.getItem('onboardingProgress');
    if (saved) {
      try {
        setProgress(JSON.parse(saved));
      } catch (error) {
        console.error('Failed to load onboarding progress:', error);
      }
    }
  }, []);

  useEffect(() => {
    // Save progress
    localStorage.setItem('onboardingProgress', JSON.stringify(progress));
  }, [progress]);

  const handleNext = () => {
    const currentStepId = steps[progress.currentStep].id;
    const nextStepIndex = progress.currentStep + 1;

    setProgress(prev => ({
      ...prev,
      completedSteps: [...prev.completedSteps, currentStepId],
      currentStep: nextStepIndex
    }));

    showNotification('Step completed', 'success');
  };

  const handleSkip = () => {
    const currentStepId = steps[progress.currentStep].id;
    const nextStepIndex = progress.currentStep + 1;

    setProgress(prev => ({
      ...prev,
      skippedSteps: [...prev.skippedSteps, currentStepId],
      currentStep: nextStepIndex
    }));

    showNotification('Step skipped', 'info');
  };

  const handleComplete = (data: any) => {
    const finalUserData = { ...progress.userData, ...data };
    
    setProgress(prev => ({
      ...prev,
      userData: finalUserData,
      isComplete: true
    }));

    // Mark onboarding as completed
    localStorage.setItem('onboardingCompleted', 'true');
    localStorage.setItem('onboardingData', JSON.stringify(finalUserData));

    showNotification('Onboarding completed! Welcome to Credence!', 'success');

    // Redirect to dashboard or trigger app initialization
    setTimeout(() => {
      window.location.reload();
    }, 2000);
  };

  const setStepData = (data: any) => {
    setProgress(prev => ({
      ...prev,
      userData: { ...prev.userData, ...data }
    }));
  };

  const currentStep = steps[progress.currentStep];
  const StepComponent = currentStep.component;

  return (
    <div className="onboarding-wizard">
      <div className="onboarding-header">
        <div className="progress-bar">
          <div className="progress-steps">
            {steps.map((step, index) => (
              <div
                key={step.id}
                className={`progress-step ${
                  index < progress.currentStep ? 'completed' :
                  index === progress.currentStep ? 'active' : 'pending'
                }`}
              >
                <div className="step-number">{index + 1}</div>
                <div className="step-info">
                  <div className="step-title">{step.title}</div>
                  <div className="step-subtitle">{step.subtitle}</div>
                </div>
              </div>
            ))}
          </div>
          <div 
            className="progress-line"
            style={{
              width: `${(progress.currentStep / (steps.length - 1)) * 100}%`
            }}
          />
        </div>
      </div>

      <div className="onboarding-content">
        <StepComponent
          onNext={handleNext}
          onSkip={currentStep.canSkip ? handleSkip : undefined}
          onComplete={handleComplete}
          stepData={progress.userData}
          setStepData={setStepData}
        />
      </div>

      <div className="onboarding-footer">
        <div className="step-counter">
          Step {progress.currentStep + 1} of {steps.length}
        </div>
        <div className="help-link">
          <a href="#" onClick={(e) => { e.preventDefault(); /* Open help */ }}>
            Need help? 📖
          </a>
        </div>
      </div>
    </div>
  );
}

export default OnboardingWizard;