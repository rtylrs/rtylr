/**
 * Platform detection utilities for Tauri vs Web
 */

export type PlatformType = 'tauri' | 'web';

export interface PlatformCapabilities {
  hasFileSystem: boolean;
  hasNativeNotifications: boolean;
  hasSystemTray: boolean;
  hasMenuBar: boolean;
  supportsDeepLinking: boolean;
  canMinimizeToTray: boolean;
}

// Check if running in Tauri desktop app
export const isTauri = (): boolean => {
  return typeof window !== 'undefined' && window.__TAURI__ !== undefined;
};

// Check if running in web browser
export const isWeb = (): boolean => {
  return !isTauri();
};

export const canUseNativeHardware = (): boolean => {
  return isTauri() || import.meta.env.MODE === 'test';
};

// Get current platform
export const getPlatform = (): PlatformType => {
  return isTauri() ? 'tauri' : 'web';
};

// Get platform capabilities
export const getPlatformCapabilities = (): PlatformCapabilities => {
  const platform = getPlatform();

  if (platform === 'tauri') {
    return {
      hasFileSystem: true,
      hasNativeNotifications: true,
      hasSystemTray: true,
      hasMenuBar: true,
      supportsDeepLinking: true,
      canMinimizeToTray: true,
    };
  }

  return {
    hasFileSystem: false,
    hasNativeNotifications: 'Notification' in window,
    hasSystemTray: false,
    hasMenuBar: false,
    supportsDeepLinking: false,
    canMinimizeToTray: false,
  };
};

// Platform-specific configuration
export const platformConfig = {
  isTauri: isTauri(),
  isWeb: isWeb(),
  platform: getPlatform(),
  capabilities: getPlatformCapabilities(),
};

// Utility to conditionally execute code based on platform
export const whenPlatform = <T>(platform: PlatformType, callback: () => T): T | undefined => {
  return getPlatform() === platform ? callback() : undefined;
};

// Utility to run different code for each platform
export const platformSwitch = <T>(handlers: {
  tauri?: () => T;
  web?: () => T;
  default?: () => T;
}): T | undefined => {
  const platform = getPlatform();
  const handler = handlers[platform] || handlers.default;
  return handler ? handler() : undefined;
};

// Declare global Tauri object for TypeScript
declare global {
  interface Window {
    __TAURI__?: unknown;
  }
}
