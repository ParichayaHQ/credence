import { Tray, Menu, nativeImage, Notification } from 'electron';
import { join } from 'path';
import { APP_NAME } from '@shared/constants';

interface TrayOptions {
  onShow: () => void;
  onQuit: () => void;
}

export class TrayManager {
  private tray: Tray | null = null;
  private options: TrayOptions;
  private notificationCount = 0;

  constructor(options: TrayOptions) {
    this.options = options;
    this.createTray();
  }

  private createTray(): void {
    try {
      const iconPath = this.getTrayIconPath();
      this.tray = new Tray(iconPath);
      
      this.tray.setToolTip(APP_NAME);
      this.setupContextMenu();
      this.setupEventHandlers();
      
      console.log('System tray created successfully');
    } catch (error) {
      console.error('Failed to create system tray:', error);
      throw error;
    }
  }

  private getTrayIconPath(): string {
    const iconDir = join(__dirname, '../../assets/icons/tray');
    
    // Use appropriate icon for platform and theme
    let iconName: string;
    
    if (process.platform === 'darwin') {
      // macOS uses template icons that adapt to dark/light theme
      iconName = 'trayTemplate.png';
    } else if (process.platform === 'win32') {
      // Windows uses ICO format
      iconName = 'tray.ico';
    } else {
      // Linux uses PNG
      iconName = 'tray.png';
    }
    
    return join(iconDir, iconName);
  }

  private setupContextMenu(): void {
    if (!this.tray) return;

    const contextMenu = Menu.buildFromTemplate([
      {
        label: 'Show Wallet',
        type: 'normal',
        click: this.options.onShow,
      },
      {
        type: 'separator',
      },
      {
        label: 'Network Status',
        type: 'submenu',
        submenu: [
          {
            label: 'Wallet Service',
            type: 'normal',
            enabled: false,
            // TODO: Update with actual status
          },
          {
            label: 'P2P Gateway',
            type: 'normal',
            enabled: false,
            // TODO: Update with actual status
          },
          {
            label: 'Scorer Service',
            type: 'normal',
            enabled: false,
            // TODO: Update with actual status
          },
        ],
      },
      {
        label: 'Recent Activity',
        type: 'submenu',
        submenu: [
          {
            label: 'No recent activity',
            type: 'normal',
            enabled: false,
          },
          // TODO: Add recent vouches, score updates, etc.
        ],
      },
      {
        type: 'separator',
      },
      {
        label: 'Settings',
        type: 'normal',
        enabled: false, // TODO: Enable when settings UI is ready
      },
      {
        label: 'Lock Wallet',
        type: 'normal',
        enabled: false, // TODO: Enable when lock/unlock is implemented
      },
      {
        type: 'separator',
      },
      {
        label: 'Quit Credence',
        type: 'normal',
        click: this.options.onQuit,
      },
    ]);

    this.tray.setContextMenu(contextMenu);
  }

  private setupEventHandlers(): void {
    if (!this.tray) return;

    // Double-click to show window
    this.tray.on('double-click', () => {
      this.options.onShow();
    });

    // Single click behavior (platform-specific)
    this.tray.on('click', () => {
      if (process.platform === 'win32') {
        // On Windows, single click shows the window
        this.options.onShow();
      }
      // On macOS, single click shows context menu (default behavior)
    });

    // Right-click shows context menu (Windows/Linux)
    this.tray.on('right-click', () => {
      if (this.tray) {
        this.tray.popUpContextMenu();
      }
    });
  }

  public updateNetworkStatus(status: {
    walletService: boolean;
    p2pGateway: boolean;
    scorer: boolean;
  }): void {
    // Update the context menu with current network status
    if (!this.tray) return;

    const contextMenu = Menu.buildFromTemplate([
      {
        label: 'Show Wallet',
        type: 'normal',
        click: this.options.onShow,
      },
      {
        type: 'separator',
      },
      {
        label: 'Network Status',
        type: 'submenu',
        submenu: [
          {
            label: `Wallet Service ${status.walletService ? '✓' : '✗'}`,
            type: 'normal',
            enabled: false,
          },
          {
            label: `P2P Gateway ${status.p2pGateway ? '✓' : '✗'}`,
            type: 'normal',
            enabled: false,
          },
          {
            label: `Scorer Service ${status.scorer ? '✓' : '✗'}`,
            type: 'normal',
            enabled: false,
          },
        ],
      },
      {
        type: 'separator',
      },
      {
        label: 'Quit Credence',
        type: 'normal',
        click: this.options.onQuit,
      },
    ]);

    this.tray.setContextMenu(contextMenu);

    // Update tray icon based on overall status
    const allGood = status.walletService && status.p2pGateway && status.scorer;
    this.setStatus(allGood ? 'connected' : 'disconnected');
  }

  public setStatus(status: 'connected' | 'disconnected' | 'syncing'): void {
    if (!this.tray) return;

    try {
      let iconPath: string;
      let tooltip: string;

      switch (status) {
        case 'connected':
          iconPath = this.getTrayIconPath(); // Default icon
          tooltip = `${APP_NAME} - Connected`;
          break;
        case 'disconnected':
          iconPath = this.getTrayIconPath(); // TODO: Use disconnected icon variant
          tooltip = `${APP_NAME} - Disconnected`;
          break;
        case 'syncing':
          iconPath = this.getTrayIconPath(); // TODO: Use syncing icon variant
          tooltip = `${APP_NAME} - Syncing`;
          break;
      }

      const icon = nativeImage.createFromPath(iconPath);
      if (!icon.isEmpty()) {
        this.tray.setImage(icon);
      }
      
      this.tray.setToolTip(tooltip);
    } catch (error) {
      console.warn('Failed to update tray status:', error);
    }
  }

  public showNotification(title: string, body: string, silent = false): void {
    if (!Notification.isSupported()) {
      console.warn('Notifications not supported on this platform');
      return;
    }

    try {
      const notification = new Notification({
        title,
        body,
        silent,
        icon: this.getTrayIconPath(),
      });

      notification.show();

      // Update badge count
      this.notificationCount++;
      this.updateBadge();

      // Handle notification click
      notification.on('click', () => {
        this.options.onShow();
        this.clearBadge();
      });
    } catch (error) {
      console.error('Failed to show notification:', error);
    }
  }

  public setBadge(count: number): void {
    this.notificationCount = Math.max(0, count);
    this.updateBadge();
  }

  public clearBadge(): void {
    this.notificationCount = 0;
    this.updateBadge();
  }

  private updateBadge(): void {
    if (!this.tray) return;

    try {
      if (this.notificationCount > 0) {
        // On macOS, we can show a badge count
        if (process.platform === 'darwin') {
          // TODO: Create badge overlay for icon
        }
        
        // Update tooltip to show notification count
        this.tray.setToolTip(`${APP_NAME} (${this.notificationCount} notifications)`);
      } else {
        this.tray.setToolTip(APP_NAME);
      }
    } catch (error) {
      console.warn('Failed to update tray badge:', error);
    }
  }

  public shouldMinimizeToTray(): boolean {
    // TODO: Make this configurable via settings
    return true; // Default to minimizing to tray
  }

  public destroy(): void {
    if (this.tray) {
      this.tray.destroy();
      this.tray = null;
      console.log('System tray destroyed');
    }
  }
}

/**
 * Utility functions for tray icon management
 */
export class TrayIconUtils {
  /**
   * Create a badge overlay for the tray icon (macOS)
   */
  static createBadgedIcon(basePath: string, count: number): Electron.NativeImage | null {
    try {
      // This would require canvas or similar to draw badge overlay
      // For now, return the base icon
      return nativeImage.createFromPath(basePath);
    } catch (error) {
      console.error('Failed to create badged icon:', error);
      return null;
    }
  }

  /**
   * Get appropriate icon size for platform
   */
  static getIconSize(): { width: number; height: number } {
    switch (process.platform) {
      case 'darwin':
        return { width: 22, height: 22 }; // macOS menu bar height
      case 'win32':
        return { width: 16, height: 16 }; // Windows system tray
      default:
        return { width: 22, height: 22 }; // Linux
    }
  }
}