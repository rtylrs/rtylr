/**
 * Local approval cache: remembers user decisions so the same tool isn't
 * re-prompted within a single chat session. Keys are `${chatId}:${tool}`.
 */

const STORAGE_KEY = 'caret-approvals';

interface ApprovalRecord {
  decision: 'approved' | 'denied';
  at: number;
}

type Store = Record<string, ApprovalRecord>;

function read(): Store {
  if (typeof localStorage === 'undefined') return {};
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return {};
    return JSON.parse(raw) as Store;
  } catch {
    return {};
  }
}

function write(store: Store): void {
  if (typeof localStorage === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(store));
  } catch {
    // ignore quota errors — approvals are best-effort
  }
}

export function rememberApproval(
  chatId: string,
  tool: string,
  decision: 'approved' | 'denied'
): void {
  const store = read();
  store[`${chatId}:${tool}`] = { decision, at: Date.now() };
  write(store);
}

export function getApproval(chatId: string, tool: string): ApprovalRecord | null {
  const store = read();
  return store[`${chatId}:${tool}`] ?? null;
}

export function clearApprovals(chatId: string): void {
  const store = read();
  for (const key of Object.keys(store)) {
    if (key.startsWith(`${chatId}:`)) delete store[key];
  }
  write(store);
}
