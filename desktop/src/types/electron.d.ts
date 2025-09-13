/**
 * Type declarations for the Electron APIs exposed via contextBridge
 * This allows TypeScript to understand the wallet APIs in the renderer process
 */

import { WalletAPI } from '../preload/wallet-api';
import { SystemAPI } from '../preload/system-api';

// EventAPI interface to match what components expect
interface EventAPI {
  onNetworkStatusChange: (callback: (status: any) => void) => () => void;
  onWalletStatusChange: (callback: (locked: boolean) => void) => () => void;
  onNewEvent: (callback: (event: any) => void) => () => void;
}

declare global {
  interface Window {
    walletAPI: WalletAPI;
    systemAPI: SystemAPI;
    eventAPI: EventAPI;
  }
}

export {};