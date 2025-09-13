import { app, BrowserWindow, Menu, ipcMain } from 'electron';
import { join } from 'path';
import { WalletService } from './wallet-service';
import { setupIpcHandlers } from './ipc-handlers';
import { createAppMenu } from './menu';
import { setupSecurity } from './security';
import { TrayManager } from './tray';
import { APP_NAME, MAIN_WINDOW_CONFIG } from '@shared/constants';

class CredenceApp {
  private mainWindow: BrowserWindow | null = null;
  private walletService: WalletService;
  private trayManager: TrayManager | null = null;
  private isDevelopment = process.env.NODE_ENV === 'development';

  constructor() {
    this.walletService = new WalletService();
    this.setupApp();
  }

  private setupApp(): void {
    // Handle app ready
    app.whenReady().then(() => {
      this.createMainWindow();
      this.setupMenu();
      this.setupTray();
      this.setupIPC();
      this.startWalletService();
    });

    // Handle all windows closed
    app.on('window-all-closed', () => {
      if (process.platform !== 'darwin') {
        this.shutdown();
      }
    });

    // Handle app activation (macOS)
    app.on('activate', () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        this.createMainWindow();
      } else if (this.mainWindow) {
        this.mainWindow.show();
      }
    });

    // Handle before quit
    app.on('before-quit', () => {
      this.shutdown();
    });

    // Security: Prevent new window creation
    app.on('web-contents-created', (_, contents) => {
      contents.setWindowOpenHandler(({ url }) => {
        console.warn('Blocked new window creation:', url);
        return { action: 'deny' };
      });
    });
  }

  private createMainWindow(): void {
    // Apply security settings
    setupSecurity();

    this.mainWindow = new BrowserWindow({
      ...MAIN_WINDOW_CONFIG,
      title: APP_NAME,
      icon: this.getAppIcon(),
      show: false, // Don't show until ready
      webPreferences: {
        nodeIntegration: false,
        contextIsolation: true,
        allowRunningInsecureContent: false,
        webSecurity: true,
        preload: join(__dirname, '../preload/index.js'),
      },
    });

    // Load the renderer
    if (this.isDevelopment) {
      this.mainWindow.loadFile(join(__dirname, '../renderer/index.html'));
      this.mainWindow.webContents.openDevTools();
    } else {
      this.mainWindow.loadFile(join(__dirname, '../renderer/index.html'));
    }

    // Show window when ready
    this.mainWindow.once('ready-to-show', () => {
      if (this.mainWindow) {
        this.mainWindow.show();
        
        if (this.isDevelopment) {
          this.mainWindow.webContents.openDevTools();
        }
      }
    });

    // Handle window closed
    this.mainWindow.on('closed', () => {
      this.mainWindow = null;
    });

    // Handle minimize to tray (optional)
    this.mainWindow.on('minimize', (event: Electron.Event) => {
      if (this.trayManager && this.trayManager.shouldMinimizeToTray()) {
        event.preventDefault();
        this.mainWindow?.hide();
      }
    });

    // Handle window focus
    this.mainWindow.on('focus', () => {
      if (this.trayManager) {
        this.trayManager.clearBadge();
      }
    });
  }

  private setupMenu(): void {
    const menu = createAppMenu({
      isDevelopment: this.isDevelopment,
      onToggleDevTools: () => {
        if (this.mainWindow) {
          this.mainWindow.webContents.toggleDevTools();
        }
      },
      onReload: () => {
        if (this.mainWindow) {
          this.mainWindow.reload();
        }
      },
      onQuit: () => {
        this.shutdown();
      },
    });
    
    Menu.setApplicationMenu(menu);
  }

  private setupTray(): void {
    try {
      this.trayManager = new TrayManager({
        onShow: () => {
          if (this.mainWindow) {
            this.mainWindow.show();
            this.mainWindow.focus();
          }
        },
        onQuit: () => {
          this.shutdown();
        },
      });
    } catch (error) {
      console.warn('Failed to create system tray:', error);
      // Tray is optional, continue without it
    }
  }

  private setupIPC(): void {
    setupIpcHandlers({
      walletService: this.walletService,
      mainWindow: this.mainWindow,
      trayManager: this.trayManager,
    });
  }

  private async startWalletService(): Promise<void> {
    try {
      console.log('Starting wallet service...');
      await this.walletService.start();
      console.log('Wallet service started successfully');
      
      // Notify renderer that service is ready
      if (this.mainWindow) {
        this.mainWindow.webContents.send('wallet-service-ready');
      }
    } catch (error) {
      console.error('Failed to start wallet service:', error);
      
      // Show error dialog to user
      if (this.mainWindow) {
        this.mainWindow.webContents.send('wallet-service-error', {
          message: 'Failed to start wallet service',
          error: error instanceof Error ? error.message : String(error),
        });
      }
    }
  }

  private async shutdown(): Promise<void> {
    console.log('Shutting down Credence Wallet...');
    
    try {
      // Stop wallet service
      if (this.walletService) {
        await this.walletService.stop();
      }
      
      // Clean up tray
      if (this.trayManager) {
        this.trayManager.destroy();
      }
      
      // Quit app
      app.quit();
    } catch (error) {
      console.error('Error during shutdown:', error);
      // Force quit even if there's an error
      app.exit(1);
    }
  }

  private getAppIcon(): string {
    const iconPath = join(__dirname, '../../assets/icons');
    
    switch (process.platform) {
      case 'win32':
        return join(iconPath, 'icon.ico');
      case 'darwin':
        return join(iconPath, 'icon.icns');
      default:
        return join(iconPath, 'icon.png');
    }
  }

  // Public methods for IPC handlers
  public getMainWindow(): BrowserWindow | null {
    return this.mainWindow;
  }

  public getWalletService(): WalletService {
    return this.walletService;
  }
}

// Create and start the app
const credenceApp = new CredenceApp();

// Handle unhandled errors
process.on('uncaughtException', (error) => {
  console.error('Uncaught exception:', error);
  app.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled rejection at:', promise, 'reason:', reason);
});

// Export for potential testing
export default credenceApp;