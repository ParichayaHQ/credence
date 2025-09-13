// Shared TypeScript types across all processes

// Define JsonWebKey interface for compatibility
export interface JsonWebKey {
  kty?: string;
  use?: string;
  key_ops?: string[];
  alg?: string;
  kid?: string;
  x5u?: string;
  x5c?: string[];
  x5t?: string;
  'x5t#S256'?: string;
  [otherMembers: string]: unknown;
}

// DID and Key Management Types
export interface KeyPair {
  id: string;
  keyType: 'Ed25519' | 'Secp256k1';
  privateKeyJWK: JsonWebKey;
  publicKeyJWK: JsonWebKey;
  createdAt: string;
  updatedAt: string;
}

export interface DIDRecord {
  did: string;
  keyId: string;
  method: string;
  document: DIDDocument;
  createdAt: string;
  updatedAt: string;
}

export interface DIDDocument {
  '@context': string[];
  id: string;
  verificationMethod: VerificationMethod[];
  authentication: string[];
  assertionMethod: string[];
  keyAgreement: string[];
  capabilityInvocation: string[];
  capabilityDelegation: string[];
}

export interface VerificationMethod {
  id: string;
  type: string;
  controller: string;
  publicKeyJwk: JsonWebKey;
}

// Verifiable Credentials Types
export interface VerifiableCredential {
  '@context': string[];
  type: string[];
  id?: string;
  issuer: string;
  issuanceDate: string;
  expirationDate?: string;
  credentialSubject: Record<string, any>;
  proof?: CredentialProof;
}

export interface CredentialProof {
  type: string;
  created: string;
  verificationMethod: string;
  proofPurpose: string;
  jws: string;
}

export interface CredentialRecord {
  id: string;
  credential: VerifiableCredential;
  metadata: CredentialMetadata;
  createdAt: string;
  updatedAt: string;
}

export interface CredentialMetadata {
  tags: string[];
  notes: string;
  isRevoked: boolean;
  revokedAt?: string;
  verified: boolean;
  verifiedAt?: string;
}

// Trust Score and Events Types
export interface TrustScore {
  did: string;
  context: 'general' | 'commerce' | 'hiring';
  score: number;
  factors: ScoreFactors;
  proofs: InclusionProof[];
  calculatedAt: string;
  expiresAt: string;
}

export interface ScoreFactors {
  kycWeight: number;
  activityWeight: number;
  vouchWeight: number;
  reportWeight: number;
  tenureWeight: number;
  diversityBonus: number;
}

export interface InclusionProof {
  type: 'inclusion' | 'consistency';
  leafHash: string;
  leafIndex: number;
  treeSize: number;
  auditPath: string[];
  signature: string;
}

export interface VouchEvent {
  type: 'vouch' | 'report';
  from: string; // DID
  to: string;   // DID
  context: 'general' | 'commerce' | 'hiring';
  epoch: string;
  payloadCID?: string;
  nonce: string;
  issuedAt: string;
  signature: string;
}

export interface EventRecord {
  id: string;
  event: VouchEvent;
  status: 'pending' | 'published' | 'confirmed' | 'failed';
  receipt?: string;
  createdAt: string;
  updatedAt: string;
}

// Wallet Configuration Types
export interface WalletConfig {
  encryptionEnabled: boolean;
  autoLockTimeout: number; // minutes
  defaultKeyType: 'Ed25519' | 'Secp256k1';
  networkEndpoints: NetworkEndpoints;
  vouchBudgets: Record<string, number>;
}

export interface NetworkEndpoints {
  p2pGateway: string;
  scorer: string;
  rulesRegistry: string;
  logNode: string;
}

// IPC Communication Types
export interface IPCRequest<T = any> {
  id: string;
  method: string;
  params: T;
  timestamp: number;
}

export interface IPCResponse<T = any> {
  id: string;
  success: boolean;
  data?: T;
  error?: string;
  timestamp: number;
}

// Error Types
export interface WalletError {
  code: string;
  message: string;
  details?: string;
  timestamp: number;
}

// API Response Types
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  timestamp: number;
}

// Application State Types
export interface AppState {
  isLocked: boolean;
  currentDID?: string;
  connectedEndpoints: NetworkStatus;
  settings: AppSettings;
}

export interface NetworkStatus {
  p2pGateway: ConnectionStatus;
  scorer: ConnectionStatus;
  rulesRegistry: ConnectionStatus;
  logNode: ConnectionStatus;
}

export interface ConnectionStatus {
  connected: boolean;
  lastSeen?: string;
  latency?: number;
  error?: string;
}

export interface AppSettings {
  theme: 'light' | 'dark' | 'system';
  language: string;
  notifications: NotificationSettings;
  privacy: PrivacySettings;
  advanced: AdvancedSettings;
}

export interface NotificationSettings {
  enabled: boolean;
  vouchReceived: boolean;
  scoreUpdated: boolean;
  networkEvents: boolean;
}

export interface PrivacySettings {
  telemetryEnabled: boolean;
  crashReporting: boolean;
  analyticsEnabled: boolean;
}

export interface AdvancedSettings {
  debugMode: boolean;
  developerMode: boolean;
  customEndpoints: boolean;
  backupReminder: number; // days
}

// Utility Types
export type Awaitable<T> = T | Promise<T>;
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
export type RequireAtLeastOne<T, Keys extends keyof T = keyof T> =
  Pick<T, Exclude<keyof T, Keys>>
  & {
      [K in Keys]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<Keys, K>>>
    }[Keys];

// Constants
export const SUPPORTED_KEY_TYPES = ['Ed25519', 'Secp256k1'] as const;
export const SUPPORTED_CONTEXTS = ['general', 'commerce', 'hiring'] as const;
export const SUPPORTED_EVENT_TYPES = ['vouch', 'report'] as const;

export type KeyType = typeof SUPPORTED_KEY_TYPES[number];
export type TrustContext = typeof SUPPORTED_CONTEXTS[number];
export type EventType = typeof SUPPORTED_EVENT_TYPES[number];