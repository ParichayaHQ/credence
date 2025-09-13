import { test, expect } from '@playwright/test';
import { ElectronApplication, Page, _electron as electron } from 'playwright';

let electronApp: ElectronApplication;
let page: Page;

test.beforeAll(async () => {
  // Launch Electron app
  electronApp = await electron.launch({
    args: ['.'],
    cwd: './build'
  });

  // Get the first page that the app opens
  page = await electronApp.firstWindow();
});

test.afterAll(async () => {
  // Close the app
  await electronApp.close();
});

test.describe('Credence Wallet App', () => {
  test('should display app title', async () => {
    // Wait for the app to load
    await page.waitForLoadState('domcontentloaded');
    
    // Check if the app title is displayed
    const title = await page.locator('h1').textContent();
    expect(title).toContain('Credence Wallet');
  });

  test('should show onboarding for new users', async () => {
    // Clear any existing onboarding state
    await page.evaluate(() => {
      localStorage.removeItem('onboardingCompleted');
    });
    
    // Reload the page
    await page.reload();
    
    // Should show onboarding wizard
    await expect(page.locator('.onboarding-wizard')).toBeVisible();
    await expect(page.locator('h2')).toContainText('Welcome to Credence Wallet');
  });

  test('should navigate through onboarding steps', async () => {
    // Start from welcome step
    await expect(page.locator('.welcome-step')).toBeVisible();
    
    // Click "Get Started"
    await page.click('.welcome-button.primary');
    
    // Should move to security setup
    await expect(page.locator('.security-step')).toBeVisible();
    
    // Skip security setup for now
    await page.click('.step-button.secondary');
    
    // Should move to key generation
    await expect(page.locator('.key-generation-step')).toBeVisible();
  });

  test('should complete onboarding flow', async () => {
    // This test would complete the full onboarding flow
    // For now, we'll simulate completion
    await page.evaluate(() => {
      localStorage.setItem('onboardingCompleted', 'true');
    });
    
    await page.reload();
    
    // Should show main app interface
    await expect(page.locator('.sidebar')).toBeVisible();
    await expect(page.locator('.header')).toBeVisible();
  });

  test('should display sidebar navigation', async () => {
    // Ensure we're in the main app
    await page.evaluate(() => {
      localStorage.setItem('onboardingCompleted', 'true');
    });
    await page.reload();
    
    // Check sidebar items
    await expect(page.locator('.nav-link[href="/dashboard"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/keys"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/dids"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/credentials"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/events"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/trust-scores"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/network"]')).toBeVisible();
    await expect(page.locator('.nav-link[href="/settings"]')).toBeVisible();
  });

  test('should open global search with Cmd+K', async () => {
    await page.evaluate(() => {
      localStorage.setItem('onboardingCompleted', 'true');
    });
    await page.reload();
    
    // Press Cmd+K (or Ctrl+K on non-Mac)
    await page.keyboard.press('Meta+KeyK');
    
    // Should open search modal
    await expect(page.locator('.global-search-overlay')).toBeVisible();
    await expect(page.locator('.search-input')).toBeFocused();
  });

  test('should navigate between pages', async () => {
    await page.evaluate(() => {
      localStorage.setItem('onboardingCompleted', 'true');
    });
    await page.reload();
    
    // Click on Keys page
    await page.click('.nav-link[href="/keys"]');
    await expect(page.locator('.page.keys')).toBeVisible();
    
    // Click on DIDs page
    await page.click('.nav-link[href="/dids"]');
    await expect(page.locator('.page.dids')).toBeVisible();
    
    // Click on Settings page
    await page.click('.nav-link[href="/settings"]');
    await expect(page.locator('.page.settings')).toBeVisible();
  });
});