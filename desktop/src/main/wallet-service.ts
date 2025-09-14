import { spawn, ChildProcess } from 'child_process';
import { join } from 'path';
import { WALLETD_CONFIG, API_ENDPOINTS } from '@shared/constants';
// Use electron's net module instead of fetch for better compatibility
import { net } from 'electron';

export interface WalletServiceConfig {
  port: number;
  host: string;
  binaryPath?: string;
  healthCheckInterval: number;
  startupTimeout: number;
  shutdownTimeout: number;
}

export class WalletService {
  private process: ChildProcess | null = null;
  private config: WalletServiceConfig;
  private healthCheckTimer: NodeJS.Timeout | null = null;
  private isStarting = false;
  private isStopping = false;
  private startupPromise: Promise<void> | null = null;

  constructor(config?: Partial<WalletServiceConfig>) {
    this.config = {
      ...WALLETD_CONFIG,
      binaryPath: this.getBinaryPath(),
      ...config,
    };
  }

  public async start(): Promise<void> {
    if (this.isStarting || this.process) {
      return this.startupPromise || Promise.resolve();
    }

    this.isStarting = true;
    this.startupPromise = this.doStart();
    
    try {
      await this.startupPromise;
    } finally {
      this.isStarting = false;
    }
  }

  private async doStart(): Promise<void> {
    console.log('Connecting to Docker-based walletd service...');

    // Skip binary check in development - use Docker service instead
    if (process.env.NODE_ENV === 'development') {
      console.log('Development mode: Connecting to Docker wallet service on port', this.config.port);
      
      // Wait for Docker service to be ready
      await this.waitForReady();
      
      // Start health checks
      this.startHealthCheck();
      return;
    }

    // Production mode - use local binary
    const binaryPath = this.config.binaryPath!;
    if (!require('fs').existsSync(binaryPath)) {
      throw new Error(`Wallet binary not found at: ${binaryPath}`);
    }

    // Spawn the Go process
    const args = [
      '--port', this.config.port.toString(),
      '--host', this.config.host,
      '--log-level', process.env.NODE_ENV === 'development' ? 'debug' : 'info',
    ];

    this.process = spawn(binaryPath, args, {
      stdio: ['pipe', 'pipe', 'pipe'],
      env: {
        ...process.env,
        WALLETD_PORT: this.config.port.toString(),
        WALLETD_HOST: this.config.host,
      },
    });

    // Handle process events
    this.process.on('error', (error) => {
      console.error('Wallet service process error:', error);
      this.process = null;
    });

    this.process.on('exit', (code, signal) => {
      console.log(`Wallet service exited with code ${code}, signal ${signal}`);
      this.process = null;
      this.stopHealthCheck();
    });

    // Log stdout/stderr in development
    if (process.env.NODE_ENV === 'development' && this.process.stdout && this.process.stderr) {
      this.process.stdout.on('data', (data) => {
        console.log(`walletd stdout: ${data}`);
      });

      this.process.stderr.on('data', (data) => {
        console.error(`walletd stderr: ${data}`);
      });
    }

    // Wait for service to be ready
    await this.waitForReady();

    // Start health checks
    this.startHealthCheck();

    console.log(`Wallet service started on ${this.config.host}:${this.config.port}`);
  }

  public async stop(): Promise<void> {
    if (this.isStopping) {
      return;
    }

    this.isStopping = true;
    console.log('Stopping wallet service...');

    try {
      this.stopHealthCheck();

      // In development mode with Docker, there's no process to kill
      if (process.env.NODE_ENV === 'development' && !this.process) {
        console.log('Development mode: Disconnected from Docker wallet service');
        return;
      }

      // Try graceful shutdown first
      if (this.process && !this.process.killed) {
        this.process.kill('SIGTERM');

        // Wait for graceful shutdown with timeout
        await new Promise<void>((resolve) => {
          const timeout = setTimeout(() => {
            // Force kill if graceful shutdown times out
            if (this.process && !this.process.killed) {
              console.warn('Wallet service did not exit gracefully, force killing...');
              this.process.kill('SIGKILL');
            }
            resolve();
          }, this.config.shutdownTimeout);

          if (this.process) {
            this.process.on('exit', () => {
              clearTimeout(timeout);
              resolve();
            });
          } else {
            clearTimeout(timeout);
            resolve();
          }
        });
      }

      this.process = null;
      console.log('Wallet service stopped');
    } finally {
      this.isStopping = false;
    }
  }

  public isRunning(): boolean {
    // In development mode with Docker, check if we're connected to the service
    if (process.env.NODE_ENV === 'development' && !this.process) {
      return this.healthCheckTimer !== null;
    }
    return this.process !== null && !this.process.killed;
  }

  public getBaseUrl(): string {
    return `http://${this.config.host}:${this.config.port}`;
  }

  public async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.getBaseUrl()}${API_ENDPOINTS.health}`, {
        method: 'GET',
      });
      
      return response.ok;
    } catch (error) {
      console.warn('Health check failed:', error);
      return false;
    }
  }

  private async waitForReady(): Promise<void> {
    const maxAttempts = Math.ceil(this.config.startupTimeout / 1000);
    let attempts = 0;

    while (attempts < maxAttempts) {
      if (!this.process || this.process.killed) {
        throw new Error('Wallet service process died during startup');
      }

      try {
        const isHealthy = await this.healthCheck();
        if (isHealthy) {
          return; // Service is ready
        }
      } catch (error) {
        // Continue trying
      }

      attempts++;
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    throw new Error(`Wallet service failed to start within ${this.config.startupTimeout}ms`);
  }

  private startHealthCheck(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer);
    }

    this.healthCheckTimer = setInterval(async () => {
      const isHealthy = await this.healthCheck();
      
      if (!isHealthy && this.process && !this.process.killed) {
        console.error('Wallet service health check failed, process may be unresponsive');
        // Could implement restart logic here
      }
    }, this.config.healthCheckInterval);
  }

  private stopHealthCheck(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer);
      this.healthCheckTimer = null;
    }
  }

  private getBinaryPath(): string {
    const isDevelopment = process.env.NODE_ENV === 'development';
    
    if (isDevelopment) {
      // In development, look for binary in the Go project root
      const goProjectRoot = join(__dirname, '../../../..');
      return join(goProjectRoot, this.getBinaryName());
    } else {
      // In production, binary is packaged with the app
      return join(process.resourcesPath, 'bin', this.getBinaryName());
    }
  }

  private getBinaryName(): string {
    switch (process.platform) {
      case 'win32':
        return 'walletd.exe';
      case 'darwin':
        return 'walletd-darwin';
      default:
        return 'walletd';
    }
  }

  // API client methods that can be used by IPC handlers
  public async makeRequest<T = any>(
    method: string,
    endpoint: string,
    data?: any,
    timeout = 10000
  ): Promise<T> {
    if (!this.isRunning()) {
      throw new Error('Wallet service is not running');
    }

    const url = `${this.getBaseUrl()}${endpoint}`;
    const options: any = {
      method: method.toUpperCase(),
      headers: {
        'Content-Type': 'application/json',
      },
    };

    if (data && method.toUpperCase() !== 'GET') {
      options.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, options);
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`HTTP ${response.status}: ${errorText}`);
      }

      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      } else {
        return await response.text() as T;
      }
    } catch (error) {
      console.error(`API request failed: ${method} ${url}`, error);
      throw error;
    }
  }
}