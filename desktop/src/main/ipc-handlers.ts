import { ipcMain, BrowserWindow } from 'electron';
import { WalletService } from './wallet-service';
import { TrayManager } from './tray';
import { IPC_CHANNELS, API_ENDPOINTS, ERROR_CODES } from '@shared/constants';
import { IPCRequest, IPCResponse, WalletError } from '@shared/types';
import { SecurityUtils } from './security';

interface IPCHandlerDependencies {
  walletService: WalletService;
  mainWindow: BrowserWindow | null;
  trayManager: TrayManager | null;
}

export function setupIpcHandlers(deps: IPCHandlerDependencies): void {
  const { walletService, mainWindow, trayManager } = deps;

  // Utility function to handle IPC requests safely
  const handleIpcRequest = async <T, R>(
    channel: string,
    handler: (params: T) => Promise<R>
  ) => {
    ipcMain.handle(channel, async (event, request: IPCRequest<T>): Promise<IPCResponse<R>> => {
      const startTime = Date.now();
      
      try {
        // Validate request structure
        if (!request || typeof request !== 'object' || !request.id || !request.method) {
          throw new Error('Invalid IPC request format');
        }

        // Log request (with sanitized data)
        if (process.env.NODE_ENV === 'development') {
          console.log(`IPC Request: ${channel}`, {
            id: request.id,
            method: request.method,
            paramsHash: SecurityUtils.hashForLogging(JSON.stringify(request.params || {})),
          });
        }

        // Execute handler
        const data = await handler(request.params);

        const response: IPCResponse<R> = {
          id: request.id,
          success: true,
          data,
          timestamp: Date.now(),
        };

        // Log successful response
        if (process.env.NODE_ENV === 'development') {
          console.log(`IPC Response: ${channel} (${Date.now() - startTime}ms)`, {
            id: request.id,
            success: true,
          });
        }

        return response;
      } catch (error) {
        const walletError: WalletError = {
          code: error instanceof Error && error.message.includes('SERVICE_UNAVAILABLE') 
            ? ERROR_CODES.SERVICE_UNAVAILABLE 
            : ERROR_CODES.UNKNOWN_ERROR,
          message: error instanceof Error ? error.message : 'Unknown error occurred',
          details: error instanceof Error ? error.stack : undefined,
          timestamp: Date.now(),
        };

        const response: IPCResponse<R> = {
          id: request.id,
          success: false,
          error: walletError.message,
          timestamp: Date.now(),
        };

        // Log error
        console.error(`IPC Error: ${channel} (${Date.now() - startTime}ms)`, {
          id: request.id,
          error: walletError,
        });

        return response;
      }
    });
  };

  // Key Management Handlers
  handleIpcRequest(IPC_CHANNELS.GENERATE_KEY, async (params: { keyType: string }) => {
    if (!params.keyType || !['Ed25519', 'Secp256k1'].includes(params.keyType)) {
      throw new Error('Invalid key type');
    }

    return await walletService.makeRequest('POST', `${API_ENDPOINTS.keys}/generate`, {
      keyType: params.keyType,
    });
  });

  handleIpcRequest(IPC_CHANNELS.LIST_KEYS, async () => {
    return await walletService.makeRequest('GET', API_ENDPOINTS.keys);
  });

  handleIpcRequest(IPC_CHANNELS.GET_KEY, async (params: { keyId: string }) => {
    if (!params.keyId) {
      throw new Error('Key ID is required');
    }

    const sanitizedKeyId = SecurityUtils.sanitizeInput(params.keyId);
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.keys}/${sanitizedKeyId}`);
  });

  handleIpcRequest(IPC_CHANNELS.DELETE_KEY, async (params: { keyId: string }) => {
    if (!params.keyId) {
      throw new Error('Key ID is required');
    }

    const sanitizedKeyId = SecurityUtils.sanitizeInput(params.keyId);
    return await walletService.makeRequest('DELETE', `${API_ENDPOINTS.keys}/${sanitizedKeyId}`);
  });

  // DID Management Handlers
  handleIpcRequest(IPC_CHANNELS.CREATE_DID, async (params: { keyId: string; method: string }) => {
    if (!params.keyId || !params.method) {
      throw new Error('Key ID and method are required');
    }

    return await walletService.makeRequest('POST', `${API_ENDPOINTS.dids}/create`, {
      keyId: SecurityUtils.sanitizeInput(params.keyId),
      method: SecurityUtils.sanitizeInput(params.method),
    });
  });

  handleIpcRequest(IPC_CHANNELS.LIST_DIDS, async () => {
    return await walletService.makeRequest('GET', API_ENDPOINTS.dids);
  });

  handleIpcRequest(IPC_CHANNELS.GET_DID, async (params: { did: string }) => {
    if (!params.did) {
      throw new Error('DID is required');
    }

    const sanitizedDid = SecurityUtils.sanitizeInput(params.did);
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.dids}/${encodeURIComponent(sanitizedDid)}`);
  });

  handleIpcRequest(IPC_CHANNELS.RESOLVE_DID, async (params: { did: string }) => {
    if (!params.did) {
      throw new Error('DID is required');
    }

    const sanitizedDid = SecurityUtils.sanitizeInput(params.did);
    return await walletService.makeRequest('POST', `${API_ENDPOINTS.dids}/resolve`, {
      did: sanitizedDid,
    });
  });

  // Credential Management Handlers
  handleIpcRequest(IPC_CHANNELS.STORE_CREDENTIAL, async (params: { credential: any; metadata?: any }) => {
    if (!params.credential) {
      throw new Error('Credential is required');
    }

    return await walletService.makeRequest('POST', API_ENDPOINTS.credentials, {
      credential: params.credential,
      metadata: params.metadata || {},
    });
  });

  handleIpcRequest(IPC_CHANNELS.LIST_CREDENTIALS, async (params?: { filter?: any }) => {
    const queryParams = params?.filter ? `?${new URLSearchParams(params.filter).toString()}` : '';
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.credentials}${queryParams}`);
  });

  handleIpcRequest(IPC_CHANNELS.GET_CREDENTIAL, async (params: { credentialId: string }) => {
    if (!params.credentialId) {
      throw new Error('Credential ID is required');
    }

    const sanitizedId = SecurityUtils.sanitizeInput(params.credentialId);
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.credentials}/${sanitizedId}`);
  });

  handleIpcRequest(IPC_CHANNELS.DELETE_CREDENTIAL, async (params: { credentialId: string }) => {
    if (!params.credentialId) {
      throw new Error('Credential ID is required');
    }

    const sanitizedId = SecurityUtils.sanitizeInput(params.credentialId);
    return await walletService.makeRequest('DELETE', `${API_ENDPOINTS.credentials}/${sanitizedId}`);
  });

  // Event Management Handlers
  handleIpcRequest(IPC_CHANNELS.CREATE_EVENT, async (params: { 
    type: string; 
    from: string; 
    to: string; 
    context: string; 
    payloadCID?: string; 
  }) => {
    const { type, from, to, context, payloadCID } = params;

    if (!type || !from || !to || !context) {
      throw new Error('Event type, from, to, and context are required');
    }

    if (!['vouch', 'report'].includes(type)) {
      throw new Error('Invalid event type');
    }

    if (!['general', 'commerce', 'hiring'].includes(context)) {
      throw new Error('Invalid context');
    }

    return await walletService.makeRequest('POST', API_ENDPOINTS.events, {
      type: SecurityUtils.sanitizeInput(type),
      from: SecurityUtils.sanitizeInput(from),
      to: SecurityUtils.sanitizeInput(to),
      context: SecurityUtils.sanitizeInput(context),
      payloadCID: payloadCID ? SecurityUtils.sanitizeInput(payloadCID) : undefined,
    });
  });

  handleIpcRequest(IPC_CHANNELS.LIST_EVENTS, async (params?: { filter?: any }) => {
    const queryParams = params?.filter ? `?${new URLSearchParams(params.filter).toString()}` : '';
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.events}${queryParams}`);
  });

  handleIpcRequest(IPC_CHANNELS.GET_EVENT, async (params: { eventId: string }) => {
    if (!params.eventId) {
      throw new Error('Event ID is required');
    }

    const sanitizedId = SecurityUtils.sanitizeInput(params.eventId);
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.events}/${sanitizedId}`);
  });

  // Trust Score Handlers
  handleIpcRequest(IPC_CHANNELS.GET_TRUST_SCORE, async (params: { did: string; context?: string }) => {
    if (!params.did) {
      throw new Error('DID is required');
    }

    const sanitizedDid = SecurityUtils.sanitizeInput(params.did);
    const queryParams = params.context ? `?context=${params.context}` : '';
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.scores}/${encodeURIComponent(sanitizedDid)}${queryParams}`);
  });

  handleIpcRequest(IPC_CHANNELS.LIST_TRUST_SCORES, async (params?: { did?: string }) => {
    const queryParams = params?.did ? `?did=${encodeURIComponent(SecurityUtils.sanitizeInput(params.did))}` : '';
    return await walletService.makeRequest('GET', `${API_ENDPOINTS.scores}${queryParams}`);
  });

  // System Operation Handlers
  handleIpcRequest(IPC_CHANNELS.LOCK_WALLET, async (params: { password: string }) => {
    if (!params.password) {
      throw new Error('Password is required');
    }

    // Don't log passwords
    return await walletService.makeRequest('POST', '/v1/wallet/lock', {
      password: params.password,
    });
  });

  handleIpcRequest(IPC_CHANNELS.UNLOCK_WALLET, async (params: { password: string }) => {
    if (!params.password) {
      throw new Error('Password is required');
    }

    // Don't log passwords
    return await walletService.makeRequest('POST', '/v1/wallet/unlock', {
      password: params.password,
    });
  });

  handleIpcRequest(IPC_CHANNELS.CHECK_LOCK_STATUS, async () => {
    return await walletService.makeRequest('GET', '/v1/wallet/status');
  });

  handleIpcRequest(IPC_CHANNELS.GET_NETWORK_STATUS, async () => {
    const isHealthy = await walletService.healthCheck();
    return {
      walletService: {
        connected: isHealthy,
        url: walletService.getBaseUrl(),
        lastChecked: new Date().toISOString(),
      },
    };
  });

  // Window Management Handlers
  ipcMain.handle(IPC_CHANNELS.MINIMIZE_WINDOW, () => {
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.minimize();
    }
  });

  ipcMain.handle(IPC_CHANNELS.CLOSE_WINDOW, () => {
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.close();
    }
  });

  ipcMain.handle(IPC_CHANNELS.TOGGLE_DEV_TOOLS, () => {
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.toggleDevTools();
    }
  });

  // Notification Handler
  ipcMain.handle(IPC_CHANNELS.SHOW_NOTIFICATION, (event, params: { 
    title: string; 
    body: string; 
    silent?: boolean; 
  }) => {
    if (trayManager) {
      trayManager.showNotification(params.title, params.body, params.silent);
    }
  });

  // App Info Handlers
  ipcMain.handle('app:get-version', () => {
    const { app } = require('electron');
    return app.getVersion();
  });

  ipcMain.handle('app:get-name', () => {
    const { app } = require('electron');
    return app.getName();
  });

  ipcMain.handle('app:is-packaged', () => {
    const { app } = require('electron');
    return app.isPackaged;
  });

  console.log('IPC handlers initialized');
}