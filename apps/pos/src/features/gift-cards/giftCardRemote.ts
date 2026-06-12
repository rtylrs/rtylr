import type { GiftCard } from '../../shared/types/pos';

type RemoteGiftCardRecord = Record<string, unknown>;

function isRecord(value: unknown): value is RemoteGiftCardRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function asTrimmedString(value: unknown): string | undefined {
  if (typeof value !== 'string' && typeof value !== 'number') return undefined;
  const normalized = String(value).trim();
  return normalized || undefined;
}

function asBalance(value: unknown): number | undefined {
  const balance = typeof value === 'number' ? value : typeof value === 'string' ? Number(value) : NaN;
  return Number.isFinite(balance) && balance >= 0 ? balance : undefined;
}

function asBoolean(value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback;
}

function asGiftCardPayload(response: unknown): RemoteGiftCardRecord | null {
  if (!isRecord(response)) return null;

  if (isRecord(response.card)) return response.card;
  if (isRecord(response.data)) return response.data;
  return response;
}

export function normalizeRemoteGiftCard(response: unknown): GiftCard | null {
  const payload = asGiftCardPayload(response);
  if (!payload) return null;

  const code = asTrimmedString(payload.code);
  const balance = asBalance(payload.balance ?? payload.current_balance ?? payload.currentBalance);
  if (!code || balance === undefined) return null;

  const id = asTrimmedString(payload.id) ?? asTrimmedString(payload.cardId) ?? code;
  const issuedAt =
    typeof payload.issuedAt === 'string'
      ? payload.issuedAt
      : typeof payload.issued_at === 'string'
        ? payload.issued_at
        : new Date().toISOString();
  const customerName =
    typeof payload.customerName === 'string'
      ? payload.customerName
      : typeof payload.customer_name === 'string'
        ? payload.customer_name
        : undefined;

  return {
    id,
    code,
    balance,
    issuedAt,
    active: asBoolean(payload.active ?? payload.isActive ?? payload.is_active, true),
    ...(customerName ? { customerName } : {}),
  };
}

export function normalizeRemoteGiftCardList(response: unknown): GiftCard[] {
  if (!isRecord(response)) return [];
  const list = response.gift_cards ?? response.cards ?? response.data;
  if (!Array.isArray(list)) return [];
  return list
    .map((entry) => normalizeRemoteGiftCard(entry))
    .filter((card): card is GiftCard => card !== null);
}
