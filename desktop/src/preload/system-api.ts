import { contextBridge, ipcRenderer } from 'electron';

// Define the System API interface
interface SystemAPI {
  // App lifecycle
  app: {
    getVersion: () => Promise<string>;
    quit: () => void;
    restart: () => void;
  };

  // File system operations
  fs: {
    selectFile: (options?: any) => Promise<string | null>;
    selectDirectory: (options?: any) => Promise<string | null>;
    saveFile: (data: any, options?: any) => Promise<string | null>;
  };

  // OS integration
  os: {
    getPlatform: () => Promise<string>;
    getSystemInfo: () => Promise<any>;
    openExternal: (url: string) => Promise<void>;
  };

  // Window controls
  window: {
    minimize: () => void;
    maximize: () => void;
    unmaximize: () => void;
    close: () => void;
    isMaximized: () => Promise<boolean>;
  };

  // Theme and appearance
  theme: {
    getTheme: () => Promise<string>;
    setTheme: (theme: 'light' | 'dark' | 'system') => Promise<void>;
    onThemeChanged: (callback: (theme: string) => void) => void;
  };

  // Notifications
  notifications: {
    show: (title: string, body: string, options?: any) => Promise<void>;
    requestPermission: () => Promise<string>;
  };

  // Auto-updater
  updater: {
    checkForUpdates: () => Promise<any>;
    downloadUpdate: () => Promise<void>;
    installUpdate: () => Promise<void>;
    onUpdateAvailable: (callback: (info: any) => void) => void;
    onUpdateDownloaded: (callback: (info: any) => void) => void;
  };

  // Event listeners
  events: {
    onNetworkStatusChange: (callback: (status: any) => void) => () => void;
    onWalletStatusChange: (callback: (locked: boolean) => void) => () => void;
    onNewEvent: (callback: (event: any) => void) => () => void;
  };
}

// Implementation of the System API
const systemAPI: SystemAPI = {
  app: {
    getVersion: () => ipcRenderer.invoke('app:getVersion'),
    quit: () => ipcRenderer.send('app:quit'),
    restart: () => ipcRenderer.send('app:restart'),
  },

  fs: {
    selectFile: (options?: any) => ipcRenderer.invoke('fs:selectFile', options),
    selectDirectory: (options?: any) => ipcRenderer.invoke('fs:selectDirectory', options),
    saveFile: (data: any, options?: any) => ipcRenderer.invoke('fs:saveFile', data, options),
  },

  os: {
    getPlatform: () => ipcRenderer.invoke('os:getPlatform'),
    getSystemInfo: () => ipcRenderer.invoke('os:getSystemInfo'),
    openExternal: (url: string) => ipcRenderer.invoke('os:openExternal', url),
  },

  window: {
    minimize: () => ipcRenderer.send('window:minimize'),
    maximize: () => ipcRenderer.send('window:maximize'),
    unmaximize: () => ipcRenderer.send('window:unmaximize'),
    close: () => ipcRenderer.send('window:close'),
    isMaximized: () => ipcRenderer.invoke('window:isMaximized'),
  },

  theme: {
    getTheme: () => ipcRenderer.invoke('theme:getTheme'),
    setTheme: (theme: 'light' | 'dark' | 'system') => ipcRenderer.invoke('theme:setTheme', theme),
    onThemeChanged: (callback: (theme: string) => void) => {
      ipcRenderer.on('theme:changed', (_event, theme) => callback(theme));
    },
  },

  notifications: {
    show: (title: string, body: string, options?: any) => 
      ipcRenderer.invoke('notifications:show', title, body, options),
    requestPermission: () => ipcRenderer.invoke('notifications:requestPermission'),
  },

  updater: {
    checkForUpdates: () => ipcRenderer.invoke('updater:checkForUpdates'),
    downloadUpdate: () => ipcRenderer.invoke('updater:downloadUpdate'),
    installUpdate: () => ipcRenderer.invoke('updater:installUpdate'),
    onUpdateAvailable: (callback: (info: any) => void) => {
      ipcRenderer.on('updater:updateAvailable', (_event, info) => callback(info));
    },
    onUpdateDownloaded: (callback: (info: any) => void) => {
      ipcRenderer.on('updater:updateDownloaded', (_event, info) => callback(info));
    },
  },

  events: {
    onNetworkStatusChange: (callback: (status: any) => void) => {
      const unsubscribe = () => {
        ipcRenderer.removeListener('network:statusChanged', callback);
      };
      ipcRenderer.on('network:statusChanged', (_event, status) => callback(status));
      return unsubscribe;
    },
    onWalletStatusChange: (callback: (locked: boolean) => void) => {
      const unsubscribe = () => {
        ipcRenderer.removeListener('wallet:statusChanged', callback);
      };
      ipcRenderer.on('wallet:statusChanged', (_event, locked) => callback(locked));
      return unsubscribe;
    },
    onNewEvent: (callback: (event: any) => void) => {
      const unsubscribe = () => {
        ipcRenderer.removeListener('event:new', callback);
      };
      ipcRenderer.on('event:new', (_event, event) => callback(event));
      return unsubscribe;
    },
  },
};

// Expose the System API to the renderer process
contextBridge.exposeInMainWorld('systemAPI', systemAPI);

// Also expose the event API separately (for backward compatibility)
contextBridge.exposeInMainWorld('eventAPI', systemAPI.events);

// Also expose types for TypeScript
export type { SystemAPI };