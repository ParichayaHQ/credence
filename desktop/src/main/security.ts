import { session } from 'electron';

/**
 * Set up security policies for the Electron app
 * Implements security best practices for a crypto wallet application
 */
export function setupSecurity(): void {
  // Set up Content Security Policy
  session.defaultSession.webRequest.onHeadersReceived((details, callback) => {
    callback({
      responseHeaders: {
        ...details.responseHeaders,
        'Content-Security-Policy': [
          [
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline'", // Needed for React dev
            "style-src 'self' 'unsafe-inline'",   // Needed for CSS-in-JS
            "img-src 'self' data: blob:",
            "font-src 'self' data:",
            "connect-src 'self' http://localhost:* ws://localhost:*", // Allow local API calls
            "worker-src 'none'",
            "object-src 'none'",
            "frame-src 'none'",
            "base-uri 'self'",
            "form-action 'self'",
          ].join('; '),
        ],
        // Additional security headers
        'X-Content-Type-Options': ['nosniff'],
        'X-Frame-Options': ['DENY'],
        'X-XSS-Protection': ['1; mode=block'],
        'Referrer-Policy': ['strict-origin-when-cross-origin'],
        'Strict-Transport-Security': ['max-age=31536000; includeSubDomains'],
        'Permissions-Policy': [
          [
            'camera=()',
            'microphone=()',
            'geolocation=()',
            'payment=()',
            'usb=()',
          ].join(', '),
        ],
      },
    });
  });

  // Block insecure protocols
  session.defaultSession.protocol.interceptHttpProtocol('http', (request, callback) => {
    // Allow localhost connections (for development and local services)
    if (request.url.startsWith('http://localhost:') || request.url.startsWith('http://127.0.0.1:')) {
      callback({ url: request.url });
      return;
    }

    // Block all other HTTP requests in production
    if (process.env.NODE_ENV === 'production') {
      console.warn('Blocked insecure HTTP request:', request.url);
      callback({ error: -3 }); // ERR_ABORTED
      return;
    }

    callback({ url: request.url });
  });

  // Clear session data on app start (optional, for enhanced privacy)
  if (process.env.NODE_ENV === 'production') {
    session.defaultSession.clearCache();
    session.defaultSession.clearStorageData({
      storages: ['cookies', 'filesystem', 'indexdb', 'localstorage', 'shadercache', 'websql', 'serviceworkers'],
    });
  }

  // Set up certificate verification (reject self-signed certs in production)
  session.defaultSession.setCertificateVerifyProc((request, callback) => {
    // In development, allow localhost with self-signed certs
    if (process.env.NODE_ENV === 'development') {
      const isLocalhost = request.hostname === 'localhost' || request.hostname === '127.0.0.1';
      if (isLocalhost) {
        callback(0); // Accept
        return;
      }
    }

    // Use default verification for all other cases
    callback(-2); // Use default verification
  });

  // Prevent permission requests that could compromise security
  session.defaultSession.setPermissionRequestHandler((webContents, permission, callback) => {
    // Deny all permission requests by default
    // Crypto wallets shouldn't need camera, microphone, location, etc.
    const allowedPermissions: string[] = [
      // Add any permissions you actually need
    ];

    const allowed = allowedPermissions.includes(permission);
    
    if (!allowed) {
      console.warn(`Denied permission request: ${permission}`);
    }
    
    callback(allowed);
  });

  // Block external navigation attempts
  session.defaultSession.webRequest.onBeforeRequest({ urls: ['*://*/*'] }, (details, callback) => {
    const url = new URL(details.url);
    
    // Allow localhost requests (for API communication)
    if (url.hostname === 'localhost' || url.hostname === '127.0.0.1') {
      callback({});
      return;
    }

    // Allow file:// protocol (for loading local resources)
    if (url.protocol === 'file:') {
      callback({});
      return;
    }

    // Allow data: URLs (for inline resources)
    if (url.protocol === 'data:') {
      callback({});
      return;
    }

    // Block all other external requests
    console.warn('Blocked external request:', details.url);
    callback({ cancel: true });
  });

  // Set up download security
  session.defaultSession.on('will-download', (event, item) => {
    // For a crypto wallet, we might want to restrict downloads
    // Allow only specific file types if needed
    const allowedExtensions = ['.json', '.txt', '.pdf']; // Backup files, etc.
    const fileName = item.getFilename();
    const fileExtension = fileName.substring(fileName.lastIndexOf('.'));

    if (!allowedExtensions.includes(fileExtension.toLowerCase())) {
      console.warn('Blocked download of suspicious file:', fileName);
      event.preventDefault();
      return;
    }

    // Set download path to user's Downloads folder
    const { app } = require('electron');
    const path = require('path');
    item.setSavePath(path.join(app.getPath('downloads'), fileName));
  });

  console.log('Security policies initialized');
}

/**
 * Additional security utilities
 */
export class SecurityUtils {
  /**
   * Sanitize user input to prevent XSS
   */
  static sanitizeInput(input: string): string {
    return input
      .replace(/[<>]/g, '') // Remove angle brackets
      .replace(/javascript:/gi, '') // Remove javascript: protocol
      .replace(/on\w+=/gi, '') // Remove event handlers
      .trim();
  }

  /**
   * Validate that a URL is safe for the wallet to access
   */
  static isUrlSafe(url: string): boolean {
    try {
      const parsed = new URL(url);
      
      // Only allow localhost and 127.0.0.1
      const allowedHosts = ['localhost', '127.0.0.1'];
      if (!allowedHosts.includes(parsed.hostname)) {
        return false;
      }

      // Only allow HTTP for local development
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        return false;
      }

      return true;
    } catch {
      return false;
    }
  }

  /**
   * Generate a secure random string for nonces, etc.
   */
  static generateSecureRandom(length: number = 32): string {
    const crypto = require('crypto');
    return crypto.randomBytes(length).toString('hex');
  }

  /**
   * Hash sensitive data for logging (without revealing the actual data)
   */
  static hashForLogging(data: string): string {
    const crypto = require('crypto');
    return crypto.createHash('sha256').update(data).digest('hex').substring(0, 8);
  }
}