// Mock Electron APIs for testing
export const ipcRenderer = {
  invoke: jest.fn(),
  send: jest.fn(),
  on: jest.fn(),
  removeAllListeners: jest.fn(),
};

export const contextBridge = {
  exposeInMainWorld: jest.fn(),
};

export const app = {
  getPath: jest.fn(),
  quit: jest.fn(),
};

export const shell = {
  openExternal: jest.fn(),
};

export const BrowserWindow = jest.fn().mockImplementation(() => ({
  loadFile: jest.fn(),
  webContents: {
    send: jest.fn(),
  },
  on: jest.fn(),
  show: jest.fn(),
  hide: jest.fn(),
  close: jest.fn(),
}));

export const Menu = {
  buildFromTemplate: jest.fn(),
  setApplicationMenu: jest.fn(),
};

export const Tray = jest.fn().mockImplementation(() => ({
  setToolTip: jest.fn(),
  setContextMenu: jest.fn(),
  on: jest.fn(),
}));

export const nativeImage = {
  createFromPath: jest.fn(),
};