// Device and platform breakdown helpers for the insights analytics feature.

export type DeviceType = 'desktop' | 'mobile' | 'tablet';

export interface DeviceBreakdown {
  type: DeviceType;
  sessions: number;
  revenue: number;
}

/**
 * totalSessions returns the total session count across all devices.
 */
export function totalSessions(devices: DeviceBreakdown[]): number {
  return devices.reduce((sum, d) => sum + d.sessions, 0);
}

/**
 * deviceSessionShare returns the session share percentage for a device type.
 */
export function deviceSessionShare(device: DeviceBreakdown, devices: DeviceBreakdown[]): number {
  const total = totalSessions(devices);
  if (total === 0) return 0;
  return Math.round((device.sessions / total) * 1000) / 10;
}

/**
 * topDeviceByRevenue returns the device type with the highest revenue.
 */
export function topDeviceByRevenue(devices: DeviceBreakdown[]): DeviceBreakdown | undefined {
  if (devices.length === 0) return undefined;
  return devices.reduce((best, d) => (d.revenue > best.revenue ? d : best));
}

/**
 * revenuePerSession returns the average revenue per session for a device.
 */
export function revenuePerSession(device: DeviceBreakdown): number {
  if (device.sessions === 0) return 0;
  return Math.round((device.revenue / device.sessions) * 100) / 100;
}

/**
 * sortDevicesByRevenue returns devices sorted by revenue descending.
 */
export function sortDevicesByRevenue(devices: DeviceBreakdown[]): DeviceBreakdown[] {
  return [...devices].sort((a, b) => b.revenue - a.revenue);
}

/**
 * mobileShare returns the combined session share of mobile and tablet devices.
 */
export function mobileShare(devices: DeviceBreakdown[]): number {
  const total = totalSessions(devices);
  if (total === 0) return 0;
  const mobileSessions = devices
    .filter((d) => d.type === 'mobile' || d.type === 'tablet')
    .reduce((sum, d) => sum + d.sessions, 0);
  return Math.round((mobileSessions / total) * 1000) / 10;
}
