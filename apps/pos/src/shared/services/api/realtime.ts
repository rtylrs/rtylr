import { RealtimeClient, createOneKitFetch } from '@voxire/api';
import { resolveApiBaseUrl, resolveWsBaseUrl } from '../server/serverConfig';

let oneKitFetch: ReturnType<typeof createOneKitFetch> | null = null;

export async function getRealtimeTicket(): Promise<string | null> {
  const baseUrl = resolveApiBaseUrl().replace(/\/$/, '');
  try {
    if (!oneKitFetch) {
      oneKitFetch = createOneKitFetch('pos');
    }
    const response = await oneKitFetch(`${baseUrl}/pos/ws-ticket`, {
      method: 'POST',
      headers: { Accept: 'application/json' },
    });
    if (!response.ok) return null;

    const body: unknown = await response.json();
    if (!body || typeof body !== 'object' || !('ticket' in body)) return null;
    const ticket = (body as { ticket?: unknown }).ticket;
    return typeof ticket === 'string' && ticket.length > 0 ? ticket : null;
  } catch {
    return null;
  }
}

function resolveWsUrl(): string {
  // Base resolves self-host → VITE_API_URL → hosted default.
  return `${resolveWsBaseUrl()}/pos/ws`;
}

export const posRealtime = new RealtimeClient({
  url: resolveWsUrl(),
  getTicket: getRealtimeTicket,
});
