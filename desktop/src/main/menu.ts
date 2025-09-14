import { Menu, MenuItemConstructorOptions, shell } from 'electron';
import { APP_NAME } from '@shared/constants';

interface MenuOptions {
  isDevelopment: boolean;
  onToggleDevTools: () => void;
  onReload: () => void;
  onQuit: () => void;
}

export function createAppMenu(options: MenuOptions): Menu {
  const { isDevelopment, onToggleDevTools, onReload, onQuit } = options;
  
  const isMac = process.platform === 'darwin';

  const template: MenuItemConstructorOptions[] = [
    // App Menu (macOS only)
    ...(isMac ? [{
      label: APP_NAME,
      submenu: [
        { role: 'about' as const },
        { type: 'separator' as const },
        { role: 'services' as const },
        { type: 'separator' as const },
        { role: 'hide' as const },
        { role: 'hideOthers' as const },
        { role: 'unhide' as const },
        { type: 'separator' as const },
        { 
          label: 'Quit',
          accelerator: 'Command+Q',
          click: onQuit,
        },
      ],
    }] : []),

    // File Menu
    {
      label: 'File',
      submenu: [
        {
          label: 'New Wallet',
          accelerator: 'CmdOrCtrl+N',
          enabled: false, // TODO: Implement when multi-wallet support is added
        },
        {
          label: 'Open Wallet',
          accelerator: 'CmdOrCtrl+O',
          enabled: false, // TODO: Implement when wallet import is added
        },
        { type: 'separator' },
        {
          label: 'Export Wallet',
          accelerator: 'CmdOrCtrl+E',
          enabled: false, // TODO: Implement when wallet export is added
        },
        {
          label: 'Import Credential',
          accelerator: 'CmdOrCtrl+I',
          enabled: false, // TODO: Implement when credential import is added
        },
        { type: 'separator' },
        ...(isMac ? [] : [
          {
            label: 'Preferences',
            accelerator: 'CmdOrCtrl+,',
            enabled: false, // TODO: Implement when settings UI is added
          },
          { type: 'separator' as const },
          {
            label: 'Exit',
            accelerator: 'CmdOrCtrl+Q',
            click: onQuit,
          },
        ]),
      ],
    },

    // Edit Menu
    {
      label: 'Edit',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        ...(isMac ? [
          { role: 'pasteAndMatchStyle' as const },
          { role: 'delete' as const },
          { role: 'selectAll' as const },
          { type: 'separator' as const },
          {
            label: 'Speech',
            submenu: [
              { role: 'startSpeaking' as const },
              { role: 'stopSpeaking' as const },
            ],
          },
        ] : [
          { role: 'delete' as const },
          { type: 'separator' as const },
          { role: 'selectAll' as const },
        ]),
      ],
    },

    // Wallet Menu
    {
      label: 'Wallet',
      submenu: [
        {
          label: 'Generate Key',
          accelerator: 'CmdOrCtrl+G',
          enabled: false, // TODO: Enable when key generation UI is ready
        },
        {
          label: 'Create DID',
          accelerator: 'CmdOrCtrl+D',
          enabled: false, // TODO: Enable when DID creation UI is ready
        },
        { type: 'separator' },
        {
          label: 'Create Vouch',
          accelerator: 'CmdOrCtrl+V',
          enabled: false, // TODO: Enable when vouch creation UI is ready
        },
        {
          label: 'View Trust Score',
          accelerator: 'CmdOrCtrl+T',
          enabled: false, // TODO: Enable when trust score UI is ready
        },
        { type: 'separator' },
        {
          label: 'Lock Wallet',
          accelerator: 'CmdOrCtrl+L',
          enabled: false, // TODO: Enable when lock/unlock is implemented
        },
        {
          label: 'Unlock Wallet',
          accelerator: 'CmdOrCtrl+U',
          enabled: false, // TODO: Enable when lock/unlock is implemented
        },
      ],
    },

    // View Menu
    {
      label: 'View',
      submenu: [
        { role: 'reload', click: onReload },
        { role: 'forceReload' },
        { role: 'toggleDevTools', click: onToggleDevTools },
        { type: 'separator' },
        { role: 'resetZoom' },
        { role: 'zoomIn' },
        { role: 'zoomOut' },
        { type: 'separator' },
        { role: 'togglefullscreen' },
      ],
    },

    // Window Menu
    {
      label: 'Window',
      submenu: [
        { role: 'minimize' },
        { role: 'close' },
        ...(isMac ? [
          { type: 'separator' as const },
          { role: 'front' as const },
          { type: 'separator' as const },
          { role: 'window' as const },
        ] : []),
      ],
    },

    // Help Menu
    {
      role: 'help',
      submenu: [
        {
          label: 'About Credence',
          click: async () => {
            // TODO: Show about dialog with app info
            console.log('Show about dialog');
          },
        },
        {
          label: 'Documentation',
          click: async () => {
            await shell.openExternal('https://docs.credence.network');
          },
        },
        {
          label: 'Community',
          click: async () => {
            await shell.openExternal('https://github.com/ParichayaHQ/credence/discussions');
          },
        },
        {
          label: 'Report Issue',
          click: async () => {
            await shell.openExternal('https://github.com/ParichayaHQ/credence/issues/new');
          },
        },
        { type: 'separator' },
        {
          label: 'Privacy Policy',
          click: async () => {
            await shell.openExternal('https://credence.network/privacy');
          },
        },
        {
          label: 'Terms of Service',
          click: async () => {
            await shell.openExternal('https://credence.network/terms');
          },
        },
      ],
    },
  ];

  // Add development-specific menu items
  if (isDevelopment) {
    const viewMenu = template.find(menu => menu.label === 'View');
    if (viewMenu && Array.isArray(viewMenu.submenu)) {
      viewMenu.submenu.push(
        { type: 'separator' },
        {
          label: 'Open Developer Tools',
          accelerator: process.platform === 'darwin' ? 'Alt+Command+I' : 'Ctrl+Shift+I',
          click: onToggleDevTools,
        },
        {
          label: 'Reload',
          accelerator: 'CmdOrCtrl+R',
          click: onReload,
        }
      );
    }

    // Add debug menu in development
    template.splice(-1, 0, {
      label: 'Debug',
      submenu: [
        {
          label: 'Show Wallet Service Logs',
          click: () => {
            console.log('TODO: Show wallet service logs');
          },
        },
        {
          label: 'Test Notification',
          click: () => {
            // TODO: Send test notification
            console.log('TODO: Test notification');
          },
        },
        {
          label: 'Clear Storage',
          click: () => {
            console.log('TODO: Clear local storage');
          },
        },
        { type: 'separator' },
        {
          label: 'Force Crash (Test)',
          click: () => {
            process.crash();
          },
        },
      ],
    });
  }

  return Menu.buildFromTemplate(template);
}

/**
 * Context menu for the renderer process
 */
export function createContextMenu(): Menu {
  return Menu.buildFromTemplate([
    { role: 'cut' },
    { role: 'copy' },
    { role: 'paste' },
    { type: 'separator' },
    { role: 'selectAll' },
    { type: 'separator' },
    {
      label: 'Inspect Element',
      click: (item, window) => {
        if (window) {
          window.webContents.inspectElement(0, 0);
        }
      },
    },
  ]);
}