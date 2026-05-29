// Geographic breakdown helpers for the insights analytics feature.

export interface GeoRegion {
  code: string;
  name: string;
  revenue: number;
  orders: number;
}

/**
 * totalGeoRevenue returns the sum of revenue across all regions.
 */
export function totalGeoRevenue(regions: GeoRegion[]): number {
  return regions.reduce((sum, r) => sum + r.revenue, 0);
}

/**
 * topRegionByRevenue returns the region with the highest revenue.
 */
export function topRegionByRevenue(regions: GeoRegion[]): GeoRegion | undefined {
  if (regions.length === 0) return undefined;
  return regions.reduce((best, r) => (r.revenue > best.revenue ? r : best));
}

/**
 * regionRevenueShare returns the revenue share percentage for a region.
 */
export function regionRevenueShare(region: GeoRegion, regions: GeoRegion[]): number {
  const total = totalGeoRevenue(regions);
  if (total === 0) return 0;
  return Math.round((region.revenue / total) * 1000) / 10;
}

/**
 * sortRegionsByRevenue returns regions sorted by revenue descending.
 */
export function sortRegionsByRevenue(regions: GeoRegion[]): GeoRegion[] {
  return [...regions].sort((a, b) => b.revenue - a.revenue);
}

/**
 * avgOrderValueByRegion returns the average order value for a region.
 */
export function avgOrderValueByRegion(region: GeoRegion): number {
  if (region.orders === 0) return 0;
  return Math.round((region.revenue / region.orders) * 100) / 100;
}

/**
 * regionCodes extracts region codes from an array of regions.
 */
export function regionCodes(regions: GeoRegion[]): string[] {
  return regions.map((r) => r.code);
}
