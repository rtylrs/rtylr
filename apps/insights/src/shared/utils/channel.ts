// Marketing channel analytics helpers for the insights feature.

export type MarketingChannel = 'organic' | 'paid' | 'email' | 'social' | 'referral' | 'direct';

export interface ChannelMetric {
  channel: MarketingChannel;
  sessions: number;
  conversions: number;
  spend: number;
  revenue: number;
}

/**
 * channelConversionRate returns the conversion percentage for a channel.
 */
export function channelConversionRate(metric: ChannelMetric): number {
  if (metric.sessions === 0) return 0;
  return Math.round((metric.conversions / metric.sessions) * 1000) / 10;
}

/**
 * channelROI returns the return on investment for a paid channel.
 * Returns 0 when spend is 0.
 */
export function channelROI(metric: ChannelMetric): number {
  if (metric.spend === 0) return 0;
  return Math.round(((metric.revenue - metric.spend) / metric.spend) * 1000) / 10;
}

/**
 * costPerConversion returns the average spend per conversion.
 */
export function costPerConversion(metric: ChannelMetric): number {
  if (metric.conversions === 0) return 0;
  return Math.round((metric.spend / metric.conversions) * 100) / 100;
}

/**
 * topChannelByRevenue returns the channel with the highest revenue.
 */
export function topChannelByRevenue(metrics: ChannelMetric[]): ChannelMetric | undefined {
  if (metrics.length === 0) return undefined;
  return metrics.reduce((best, m) => (m.revenue > best.revenue ? m : best));
}

/**
 * sortChannelsByRevenue returns metrics sorted by revenue descending.
 */
export function sortChannelsByRevenue(metrics: ChannelMetric[]): ChannelMetric[] {
  return [...metrics].sort((a, b) => b.revenue - a.revenue);
}

/**
 * channelLabel returns a display-friendly label for a channel.
 */
export function channelLabel(channel: MarketingChannel | string): string {
  const labels: Record<string, string> = {
    organic: 'Organic Search',
    paid: 'Paid Ads',
    email: 'Email',
    social: 'Social Media',
    referral: 'Referral',
    direct: 'Direct',
  };
  return labels[channel] ?? channel;
}
