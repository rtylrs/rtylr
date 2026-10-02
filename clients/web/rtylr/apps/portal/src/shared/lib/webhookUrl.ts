export const SUPPORTED_WEBHOOK_EVENTS = ['order_created'] as const;

export function isSupportedWebhookEvent(value: string): boolean {
  return SUPPORTED_WEBHOOK_EVENTS.some((event) => event === value);
}

export function normalizeWebhookUrl(value: string): string | null {
  const normalized = value.trim();
  if (!normalized) return null;

  try {
    const parsed = new URL(normalized);
    if ((parsed.protocol !== 'http:' && parsed.protocol !== 'https:') || !parsed.hostname) {
      return null;
    }
    return normalized;
  } catch {
    return null;
  }
}
