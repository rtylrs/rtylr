import { useState, useEffect } from 'react';
import {
  isTauri,
  isWeb,
  getPlatform,
  getPlatformCapabilities,
  platformConfig,
  PlatformType,
  PlatformCapabilities,
} from '../utils/platform';

export interface PlatformInfo {
  isTauri: boolean;
  isWeb: boolean;
  platform: PlatformType;
  capabilities: PlatformCapabilities;
}

/**
 * React hook for platform detection
 * @returns PlatformInfo object with current platform details
 */
export const usePlatform = (): PlatformInfo => {
  const [platformInfo, setPlatformInfo] = useState<PlatformInfo>({
    isTauri: false,
    isWeb: true,
    platform: 'web',
    capabilities: {
      hasFileSystem: false,
      hasNativeNotifications: false,
      hasSystemTray: false,
      hasMenuBar: false,
      supportsDeepLinking: false,
      canMinimizeToTray: false,
    },
  });

  useEffect(() => {
    // Update platform info after component mounts
    setPlatformInfo({
      isTauri: isTauri(),
      isWeb: isWeb(),
      platform: getPlatform(),
      capabilities: getPlatformCapabilities(),
    });
  }, []);

  return platformInfo;
};

// Export static platform config for use outside of React components
export { platformConfig };
