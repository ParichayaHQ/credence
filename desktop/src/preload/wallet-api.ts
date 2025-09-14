import { contextBridge, ipcRenderer } from 'electron';

// Define the API interface that will be exposed to the renderer
interface WalletAPI {
  // System operations
  system: {
    getNetworkStatus: () => Promise<any>;
    lock: (password: string) => Promise<void>;
    unlock: (password: string) => Promise<void>;
    checkLockStatus: () => Promise<any>;
  };

  // Notifications
  notifications: {
    show: (title: string, body: string, silent?: boolean) => Promise<void>;
  };

  // Key management
  keys: {
    list: () => Promise<{ data: any[] }>;
    get: (keyId: string) => Promise<{ data: any }>;
    generate: (keyType: string) => Promise<{ data: any }>;
    delete: (keyId: string) => Promise<void>;
  };

  // DID management
  dids: {
    list: () => Promise<{ data: any[] }>;
    get: (did: string) => Promise<{ data: any }>;
    create: (keyId: string, method: string) => Promise<{ data: any }>;
    resolve: (did: string) => Promise<{ data: any }>;
  };

  // Credential management
  credentials: {
    list: (filter?: any) => Promise<{ data: any[] }>;
    get: (credentialId: string) => Promise<{ data: any }>;
    store: (credential: any, metadata?: any) => Promise<{ data: any }>;
    delete: (credentialId: string) => Promise<void>;
  };

  // Event management
  events: {
    list: (filter?: any) => Promise<{ data: any[] }>;
    create: (type: string, from: string, to: string, context: string, payloadCID?: string) => Promise<{ data: any }>;
    getRecent: (limit: number) => Promise<{ data: any[] }>;
  };

  // Trust scores
  trustScores: {
    list: (did?: string) => Promise<{ data: any[] }>;
    get: (did: string, context?: string) => Promise<{ data: any }>;
  };

  // Network operations
  network: {
    getHealth: () => Promise<{ data: any }>;
    getPeers: () => Promise<{ data: any[] }>;
    getStats: () => Promise<{ data: any }>;
  };

  // Checkpoints
  checkpoints: {
    list: () => Promise<{ data: any[] }>;
    latest: () => Promise<{ data: any }>;
    verify: (epoch: string) => Promise<{ data: any }>;
  };

  // Rules
  rules: {
    getActive: () => Promise<{ data: any }>;
    getUpdates: () => Promise<{ data: any[] }>;
  };

  // Window controls
  window: {
    minimize: () => void;
    close: () => void;
  };
}

// Implementation of the API that calls the main process via IPC
const walletAPI: WalletAPI = {
  system: {
    getNetworkStatus: () => ipcRenderer.invoke('system:getNetworkStatus'),
    lock: (password: string) => ipcRenderer.invoke('system:lock', password),
    unlock: (password: string) => ipcRenderer.invoke('system:unlock', password),
    checkLockStatus: () => ipcRenderer.invoke('system:checkLockStatus'),
  },

  notifications: {
    show: (title: string, body: string, silent?: boolean) => 
      ipcRenderer.invoke('notifications:show', title, body, silent),
  },

  keys: {
    list: () => ipcRenderer.invoke('keys:list'),
    get: (keyId: string) => ipcRenderer.invoke('keys:get', keyId),
    generate: (keyType: string) => ipcRenderer.invoke('keys:generate', keyType),
    delete: (keyId: string) => ipcRenderer.invoke('keys:delete', keyId),
  },

  dids: {
    list: () => ipcRenderer.invoke('dids:list'),
    get: (did: string) => ipcRenderer.invoke('dids:get', did),
    create: (keyId: string, method: string) => ipcRenderer.invoke('dids:create', keyId, method),
    resolve: (did: string) => ipcRenderer.invoke('dids:resolve', did),
  },

  credentials: {
    list: (filter?: any) => ipcRenderer.invoke('credentials:list', filter),
    get: (credentialId: string) => ipcRenderer.invoke('credentials:get', credentialId),
    store: (credential: any, metadata?: any) => ipcRenderer.invoke('credentials:store', credential, metadata),
    delete: (credentialId: string) => ipcRenderer.invoke('credentials:delete', credentialId),
  },

  events: {
    list: (filter?: any) => ipcRenderer.invoke('events:list', filter),
    create: (type: string, from: string, to: string, context: string, payloadCID?: string) => 
      ipcRenderer.invoke('events:create', type, from, to, context, payloadCID),
    getRecent: (limit: number) => ipcRenderer.invoke('events:getRecent', limit),
  },

  trustScores: {
    list: (did?: string) => ipcRenderer.invoke('trustScores:list', did),
    get: (did: string, context?: string) => ipcRenderer.invoke('trustScores:get', did, context),
  },

  network: {
    getHealth: () => ipcRenderer.invoke('network:getHealth'),
    getPeers: () => ipcRenderer.invoke('network:getPeers'),
    getStats: () => ipcRenderer.invoke('network:getStats'),
  },

  checkpoints: {
    list: () => ipcRenderer.invoke('checkpoints:list'),
    latest: () => ipcRenderer.invoke('checkpoints:latest'),
    verify: (epoch: string) => ipcRenderer.invoke('checkpoints:verify', epoch),
  },

  rules: {
    getActive: () => ipcRenderer.invoke('rules:getActive'),
    getUpdates: () => ipcRenderer.invoke('rules:getUpdates'),
  },

  window: {
    minimize: () => ipcRenderer.send('window:minimize'),
    close: () => ipcRenderer.send('window:close'),
  },
};

// Expose the API to the renderer process
contextBridge.exposeInMainWorld('walletAPI', walletAPI);

// Also expose types for TypeScript
export type { WalletAPI };