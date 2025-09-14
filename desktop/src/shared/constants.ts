// Application constants shared across all processes

export const APP_NAME = 'Credence Wallet';
export const APP_VERSION = '0.1.0';
export const APP_ID = 'network.credence.wallet';

// Window configuration
export const MAIN_WINDOW_CONFIG = {
  width: 1200,
  height: 800,
  minWidth: 800,
  minHeight: 600,
} as const;

// Go service configuration
export const WALLETD_CONFIG = {
  port: 8084,
  host: '127.0.0.1', 
  healthCheckInterval: 5000, // ms
  startupTimeout: 30000, // ms
  shutdownTimeout: 10000, // ms
} as const;

// API endpoints
export const API_ENDPOINTS = {
  keys: '/v1/keys',
  dids: '/v1/dids',
  credentials: '/v1/credentials',
  presentations: '/v1/presentations',
  events: '/v1/events',
  scores: '/v1/scores',
  health: '/health',
} as const;

// IPC channels
export const IPC_CHANNELS = {
  // Wallet operations
  GENERATE_KEY: 'wallet:generateKey',
  LIST_KEYS: 'wallet:listKeys',
  GET_KEY: 'wallet:getKey',
  DELETE_KEY: 'wallet:deleteKey',
  
  // DID operations
  CREATE_DID: 'wallet:createDID',
  LIST_DIDS: 'wallet:listDIDs',
  GET_DID: 'wallet:getDID',
  RESOLVE_DID: 'wallet:resolveDID',
  
  // Credential operations
  STORE_CREDENTIAL: 'wallet:storeCredential',
  LIST_CREDENTIALS: 'wallet:listCredentials',
  GET_CREDENTIAL: 'wallet:getCredential',
  DELETE_CREDENTIAL: 'wallet:deleteCredential',
  
  // Event operations
  CREATE_EVENT: 'wallet:createEvent',
  LIST_EVENTS: 'wallet:listEvents',
  GET_EVENT: 'wallet:getEvent',
  
  // Trust score operations
  GET_TRUST_SCORE: 'wallet:getTrustScore',
  LIST_TRUST_SCORES: 'wallet:listTrustScores',
  
  // System operations
  LOCK_WALLET: 'system:lockWallet',
  UNLOCK_WALLET: 'system:unlockWallet',
  CHECK_LOCK_STATUS: 'system:checkLockStatus',
  GET_NETWORK_STATUS: 'system:getNetworkStatus',
  
  // Settings
  GET_SETTINGS: 'settings:get',
  UPDATE_SETTINGS: 'settings:update',
  
  // Notifications
  SHOW_NOTIFICATION: 'notification:show',
  
  // Window management
  MINIMIZE_WINDOW: 'window:minimize',
  CLOSE_WINDOW: 'window:close',
  TOGGLE_DEV_TOOLS: 'window:toggleDevTools',
} as const;

// Error codes
export const ERROR_CODES = {
  // General errors
  UNKNOWN_ERROR: 'UNKNOWN_ERROR',
  INVALID_PARAMS: 'INVALID_PARAMS',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
  TIMEOUT: 'TIMEOUT',
  
  // Wallet errors
  WALLET_LOCKED: 'WALLET_LOCKED',
  INVALID_PASSWORD: 'INVALID_PASSWORD',
  KEY_NOT_FOUND: 'KEY_NOT_FOUND',
  INVALID_KEY_TYPE: 'INVALID_KEY_TYPE',
  CRYPTO_ERROR: 'CRYPTO_ERROR',
  
  // DID errors
  DID_NOT_FOUND: 'DID_NOT_FOUND',
  INVALID_DID: 'INVALID_DID',
  DID_RESOLUTION_FAILED: 'DID_RESOLUTION_FAILED',
  
  // Credential errors
  CREDENTIAL_NOT_FOUND: 'CREDENTIAL_NOT_FOUND',
  INVALID_CREDENTIAL: 'INVALID_CREDENTIAL',
  CREDENTIAL_VERIFICATION_FAILED: 'CREDENTIAL_VERIFICATION_FAILED',
  
  // Network errors
  NETWORK_ERROR: 'NETWORK_ERROR',
  CONNECTION_FAILED: 'CONNECTION_FAILED',
  REQUEST_FAILED: 'REQUEST_FAILED',
  
  // Event errors
  EVENT_NOT_FOUND: 'EVENT_NOT_FOUND',
  INVALID_EVENT: 'INVALID_EVENT',
  BUDGET_EXCEEDED: 'BUDGET_EXCEEDED',
  EPOCH_INVALID: 'EPOCH_INVALID',
} as const;

// Default settings
export const DEFAULT_SETTINGS = {
  theme: 'system' as const,
  language: 'en',
  notifications: {
    enabled: true,
    vouchReceived: true,
    scoreUpdated: true,
    networkEvents: false,
  },
  privacy: {
    telemetryEnabled: false,
    crashReporting: true,
    analyticsEnabled: false,
  },
  advanced: {
    debugMode: false,
    developerMode: false,
    customEndpoints: false,
    backupReminder: 30, // days
  },
} as const;

// Default wallet configuration
export const DEFAULT_WALLET_CONFIG = {
  encryptionEnabled: true,
  autoLockTimeout: 15, // minutes
  defaultKeyType: 'Ed25519' as const,
  networkEndpoints: {
    p2pGateway: 'http://localhost:8081',
    scorer: 'http://localhost:8082',
    rulesRegistry: 'http://localhost:8084',
    logNode: 'http://localhost:8085',
  },
  vouchBudgets: {
    general: 5,
    commerce: 3,
    hiring: 2,
  },
} as const;

// Trust score thresholds
export const TRUST_SCORE_THRESHOLDS = {
  LOW: 20,
  MEDIUM: 50,
  HIGH: 80,
  EXCELLENT: 95,
} as const;

// UI Constants
export const UI_CONSTANTS = {
  SIDEBAR_WIDTH: 250,
  HEADER_HEIGHT: 60,
  FOOTER_HEIGHT: 40,
  CARD_BORDER_RADIUS: 8,
  ANIMATION_DURATION: 200,
} as const;

// File extensions
export const SUPPORTED_IMPORT_FORMATS = [
  '.json',
  '.jwt',
  '.jwk',
  '.pem',
] as const;

export const SUPPORTED_EXPORT_FORMATS = [
  'json',
  'jwt',
  'jwk',
  'pem',
] as const;

// Validation patterns
export const VALIDATION_PATTERNS = {
  DID: /^did:[a-z0-9]+:[a-zA-Z0-9._%-]*[a-zA-Z0-9._%-]+$/,
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  URL: /^https?:\/\/.+/,
  HEX: /^[0-9a-fA-F]+$/,
  BASE64: /^[A-Za-z0-9+/]*={0,2}$/,
} as const;

// Time constants
export const TIME_CONSTANTS = {
  SECOND: 1000,
  MINUTE: 60 * 1000,
  HOUR: 60 * 60 * 1000,
  DAY: 24 * 60 * 60 * 1000,
  WEEK: 7 * 24 * 60 * 60 * 1000,
  MONTH: 30 * 24 * 60 * 60 * 1000,
} as const;