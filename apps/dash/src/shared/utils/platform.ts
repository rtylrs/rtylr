/**
 * Platform detection utilities for Web
 */

export type PlatformType = 'web';

export interface PlatformCapabilities {
  hasFileSystem: boolean;
  hasNativeNotifications: boolean;
  hasSystemTray: boolean;
  hasMenuBar: boolean;
  supportsDeepLinking: boolean;
  canMinimizeToTray: boolean;
}

// Check if running in Tauri desktop app (always false in web-only version)
export const isTauri = (): boolean => {
  return false;
};

// Check if running in web browser (always true in web-only version)
export const isWeb = (): boolean => {
  return true;
};

// Get current platform (always web in web-only version)
export const getPlatform = (): PlatformType => {
  return 'web';
};

// Get platform capabilities (web-only)
export const getPlatformCapabilities = (): PlatformCapabilities => {
  return {
    hasFileSystem: false,
    hasNativeNotifications: 'Notification' in window,
    hasSystemTray: false,
    hasMenuBar: false,
    supportsDeepLinking: false,
    canMinimizeToTray: false,
  };
};

// Platform-specific configuration (web-only)
export const platformConfig = {
  isWeb: isWeb(),
  platform: getPlatform(),
  capabilities: getPlatformCapabilities(),
};

// Utility to conditionally execute code based on platform
export const whenPlatform = <T>(platform: PlatformType, callback: () => T): T | undefined => {
  return getPlatform() === platform ? callback() : undefined;
};

// Utility to run different code for each platform (web-only version)
export const platformSwitch = <T>(handlers: {
  web?: () => T;
  default?: () => T;
}): T | undefined => {
  const handler = handlers.web || handlers.default;
  return handler ? handler() : undefined;
};
