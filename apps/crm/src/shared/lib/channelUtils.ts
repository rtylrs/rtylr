// Channel helpers for the CRM messaging features.

export type Channel = 'email' | 'sms' | 'push';

const CHANNEL_LABELS: Record<Channel, string> = {
  email: 'Email',
  sms: 'SMS',
  push: 'Push',
};

const CHANNEL_ICONS: Record<Channel, string> = {
  email: 'mail',
  sms: 'message-circle',
  push: 'bell',
};

const MAX_LENGTH: Record<Channel, number> = {
  email: 10000,
  sms: 160,
  push: 256,
};

/**
 * channelLabel returns a human-readable label for a channel.
 */
export function channelLabel(channel: Channel | string): string {
  return (CHANNEL_LABELS as Record<string, string>)[channel] ?? channel;
}

/**
 * channelIcon returns an icon name for a channel.
 */
export function channelIcon(channel: Channel | string): string {
  return (CHANNEL_ICONS as Record<string, string>)[channel] ?? 'circle';
}

/**
 * maxMessageLength returns the character limit for a channel.
 */
export function maxMessageLength(channel: Channel | string): number {
  return (MAX_LENGTH as Record<string, number>)[channel] ?? 10000;
}

/**
 * isWithinLimit returns true when the message fits within the channel's limit.
 */
export function isWithinLimit(message: string, channel: Channel | string): boolean {
  return message.length <= maxMessageLength(channel);
}

/**
 * smsSegments returns the number of 160-character SMS segments needed.
 */
export function smsSegments(message: string): number {
  return Math.ceil(message.length / 160) || 1;
}

/**
 * allChannels returns all supported channel types.
 */
export function allChannels(): Channel[] {
  return ['email', 'sms', 'push'];
}
