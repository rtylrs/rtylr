import { ResultAsync, unwrap } from '@voxire/api';
import type { ApiError, SafeApi } from '@voxire/api';
import type { AxiosRequestConfig } from 'axios';
import {
  db,
  type ApiMutation,
  type ApiMutationMethod,
  type ApiRequestMethod,
  type ApiRequestStatus,
} from '../db';

type RecordLike = Record<string, unknown>;

const NON_QUEUEABLE_MUTATION_PATTERNS = [
  /\/heartbeat$/,
  /\/verify(?:\/|$|-)/,
  /\/verify-pin$/,
  /\/manager\/verify$/,
  /\/access\//,
  /\/users\//,
];

const TABLES_URL = '/api/pos/table/v1/tables';
const ORDERS_URL = '/api/pos/orders/v1/orders';
const HELD_ORDERS_URL = '/api/pos/orders/holds/v1/orders/held';

const COLLECTION_KEYS: Record<string, string[]> = {
  [TABLES_URL]: ['tables', 'data'],
  [ORDERS_URL]: ['orders', 'data'],
  [HELD_ORDERS_URL]: ['orders', 'heldOrders', 'data'],
  '/api/pos/menu-item/v1/menu-items': ['items', 'products', 'data'],
  '/api/pos/customer/v1/customers': ['customers', 'data'],
  '/api/pos/gift-card/v1/gift-cards': ['giftCards', 'cards', 'data'],
  '/api/pos/modifier/v1/modifiers': ['modifiers', 'groups', 'data'],
  '/api/pos/combo/v1/combos': ['combos', 'data'],
  '/api/pos/stock/v1/inventory': ['items', 'inventory', 'data'],

  // Keep old snapshots readable until clients naturally refresh them.
  '/pos/tables': ['tables', 'data'],
  '/pos/orders': ['orders', 'data'],
  '/pos/orders/held': ['orders', 'heldOrders', 'data'],
  '/pos/menu-items': ['items', 'products', 'data'],
  '/pos/customers': ['customers', 'data'],
  '/pos/gift-cards': ['giftCards', 'cards', 'data'],
  '/pos/modifiers': ['modifiers', 'groups', 'data'],
  '/pos/combos': ['combos', 'data'],
  '/pos/stock/inventory': ['items', 'inventory', 'data'],
};

let syncStarted = false;
let syncing = false;

export function createLocalFirstSafeApi(remote: SafeApi): SafeApi {
  return {
    get: <T>(url: string, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(localFirstGet<T>(remote, url, config), toApiError),
    post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(localFirstMutation<T>(remote, 'post', url, data, config), toApiError),
    put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(localFirstMutation<T>(remote, 'put', url, data, config), toApiError),
    patch: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(localFirstMutation<T>(remote, 'patch', url, data, config), toApiError),
    delete: <T>(url: string, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(localFirstDelete<T>(remote, url, config), toApiError),
  };
}

export function startLocalFirstSync(remote: SafeApi): void {
  if (syncStarted || typeof window === 'undefined') return;
  syncStarted = true;

  const sync = (): void => {
    void syncPendingApiMutations(remote);
  };

  window.addEventListener('online', sync);
  window.addEventListener('rtylr-local-first-sync', sync);
  if (isOnline()) {
    window.setTimeout(sync, 500);
  }
}

export async function syncPendingApiMutations(remote: SafeApi): Promise<void> {
  if (syncing || !isOnline()) return;
  syncing = true;

  try {
    const queued = await db.apiMutations
      .where('status')
      .anyOf('pending', 'failed')
      .sortBy('createdAt');

    for (const mutation of queued) {
      await db.apiMutations.update(mutation.id, {
        status: 'syncing',
        retries: mutation.retries + 1,
        updatedAt: Date.now(),
      });

      try {
        const response = await callRemote<unknown>(
          remote,
          mutation.method,
          mutation.url,
          mutation.requestBody,
          configFromParams(mutation.params)
        );
        await applySyncedMutationResult(mutation, response);
        await db.apiMutations.delete(mutation.id);
        await logApiRequest({
          method: mutation.method,
          url: mutation.url,
          params: mutation.params,
          requestBody: mutation.requestBody,
          responseBody: response,
          status: 'ok',
        });
      } catch (error) {
        const apiError = toApiError(error);
        await db.apiMutations.update(mutation.id, {
          status: 'failed',
          lastError: apiError.message,
          updatedAt: Date.now(),
        });
        await logApiRequest({
          method: mutation.method,
          url: mutation.url,
          params: mutation.params,
          requestBody: mutation.requestBody,
          status: 'failed',
          error: apiError.message,
        });

        if (apiError.code === 'UNAUTHORIZED' || apiError.code === 'FORBIDDEN') {
          break;
        }
      }
    }
  } finally {
    syncing = false;
  }
}

async function localFirstGet<T>(
  remote: SafeApi,
  url: string,
  config?: AxiosRequestConfig
): Promise<T> {
  const params = storageSafe(config?.params);
  const key = requestKey('get', url, params);

  if (!isOnline()) {
    const cached = await db.apiSnapshots.get(key);
    if (cached) {
      await logApiRequest({
        method: 'get',
        url,
        params,
        responseBody: cached.responseBody,
        status: 'cached',
      });
      return cached.responseBody as T;
    }
    throw networkApiError('Offline and no local copy exists yet.');
  }

  try {
    const response = await unwrap(remote.get<T>(url, config));
    await cacheResponse('get', url, params, undefined, response);
    await logApiRequest({ method: 'get', url, params, responseBody: response, status: 'ok' });
    return response;
  } catch (error) {
    const apiError = toApiError(error);
    if (isRecoverableNetworkError(apiError)) {
      const cached = await db.apiSnapshots.get(key);
      if (cached) {
        await logApiRequest({
          method: 'get',
          url,
          params,
          responseBody: cached.responseBody,
          status: 'cached',
        });
        return cached.responseBody as T;
      }
    }

    await logApiRequest({ method: 'get', url, params, status: 'failed', error: apiError.message });
    throw apiError;
  }
}

async function localFirstMutation<T>(
  remote: SafeApi,
  method: ApiMutationMethod,
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig
): Promise<T> {
  const params = storageSafe(config?.params);
  const requestBody = storageSafe(data);

  if (!isOnline()) {
    if (!isQueueableMutation(url)) {
      throw networkApiError('This action needs a connection.');
    }
    return enqueueMutation<T>(method, url, requestBody, params);
  }

  try {
    const response = await callRemote<T>(remote, method, url, data, config);
    await cacheResponse(method, url, params, requestBody, response);
    await applyMutationToSnapshots(method, url, requestBody, response);
    await logApiRequest({
      method,
      url,
      params,
      requestBody,
      responseBody: response,
      status: 'ok',
    });
    void syncPendingApiMutations(remote);
    return response;
  } catch (error) {
    const apiError = toApiError(error);
    if (!isRecoverableNetworkError(apiError) || !isQueueableMutation(url)) {
      await logApiRequest({
        method,
        url,
        params,
        requestBody,
        status: 'failed',
        error: apiError.message,
      });
      throw apiError;
    }

    return enqueueMutation<T>(method, url, requestBody, params, apiError.message);
  }
}

async function localFirstDelete<T>(
  remote: SafeApi,
  url: string,
  config?: AxiosRequestConfig
): Promise<T> {
  return localFirstMutation<T>(remote, 'delete', url, undefined, config);
}

async function enqueueMutation<T>(
  method: ApiMutationMethod,
  url: string,
  requestBody?: unknown,
  params?: unknown,
  lastError?: string
): Promise<T> {
  const optimisticResponse = buildOptimisticResponse(method, url, requestBody);
  const now = Date.now();
  const mutation: ApiMutation = {
    id: `api_${now}_${Math.random().toString(36).slice(2, 8)}`,
    method,
    url,
    params,
    requestBody,
    optimisticResponse,
    status: 'pending',
    retries: 0,
    lastError,
    createdAt: now,
    updatedAt: now,
  };

  await db.apiMutations.put(mutation);
  await applyMutationToSnapshots(method, url, requestBody, optimisticResponse);
  await cacheResponse(method, url, params, requestBody, optimisticResponse);
  await logApiRequest({
    method,
    url,
    params,
    requestBody,
    responseBody: optimisticResponse,
    status: 'queued',
  });
  dispatchLocalFirstEvent('rtylr-local-first-queued');
  return optimisticResponse as T;
}

async function callRemote<T>(
  remote: SafeApi,
  method: ApiMutationMethod,
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig
): Promise<T> {
  if (method === 'delete') {
    return unwrap(remote.delete<T>(url, config));
  }
  return unwrap(remote[method]<T>(url, data, config));
}

async function cacheResponse(
  method: ApiRequestMethod,
  url: string,
  params: unknown,
  requestBody: unknown,
  responseBody: unknown
): Promise<void> {
  await db.apiSnapshots.put({
    key: requestKey(method, url, params, requestBody),
    method,
    url,
    params,
    requestBody,
    responseBody,
    updatedAt: Date.now(),
  });
}

async function applySyncedMutationResult(mutation: ApiMutation, response: unknown): Promise<void> {
  const optimisticId = idOf(mutation.optimisticResponse);
  const remoteId = idOf(response);

  if (isOrdersCollectionUrl(mutation.url) && optimisticId && remoteId && optimisticId !== remoteId) {
    const merged = mergeEntity(asRecord(mutation.optimisticResponse) ?? {}, response, false);
    await updateCollectionSnapshots(ORDERS_URL, ['orders', 'data'], (existing) =>
      upsertById(removeById(existing, optimisticId), merged)
    );
    await cacheResponse('get', `${ORDERS_URL}/${remoteId}`, undefined, undefined, merged);
    return;
  }

  if (isTablesCollectionUrl(mutation.url) && optimisticId && remoteId && optimisticId !== remoteId) {
    const merged = mergeEntity(asRecord(mutation.optimisticResponse) ?? {}, response, false);
    await updateCollectionSnapshots(TABLES_URL, ['tables', 'data'], (existing) =>
      upsertById(removeById(existing, optimisticId), merged)
    );
    return;
  }

  await applyMutationToSnapshots(mutation.method, mutation.url, mutation.requestBody, response);
}

async function applyMutationToSnapshots(
  method: ApiMutationMethod,
  url: string,
  requestBody: unknown,
  responseBody: unknown
): Promise<void> {
  if (isOrdersCollectionUrl(url) && method === 'post') {
    const order = mergeEntity(
      buildOptimisticOrder(requestBody),
      responseBody,
      Boolean(asRecord(responseBody)?.pendingSync)
    );
    await updateCollectionSnapshots(ORDERS_URL, ['orders', 'data'], (existing) =>
      upsertById(existing, order)
    );
    const orderId = idOf(order);
    if (orderId) {
      await cacheResponse('get', `${ORDERS_URL}/${orderId}`, undefined, undefined, order);
    }
    return;
  }

  if (isTablesCollectionUrl(url) && method === 'post') {
    const table = mergeEntity(
      buildOptimisticTable(requestBody),
      responseBody,
      Boolean(asRecord(responseBody)?.pendingSync)
    );
    await updateCollectionSnapshots(TABLES_URL, ['tables', 'data'], (existing) =>
      upsertById(existing, table)
    );
    return;
  }

  const tableStatusMatch = tableStatusUrlMatch(url);
  if (tableStatusMatch && (method === 'put' || method === 'patch')) {
    const id = tableStatusMatch[1];
    const status = asRecord(requestBody)?.status;
    await updateCollectionSnapshots(TABLES_URL, ['tables', 'data'], (existing) =>
      existing.map((entry) =>
        sameId(entry, id) ? { ...(asRecord(entry) ?? {}), id: idValue(entry, id), status } : entry
      )
    );
    return;
  }

  const tableMatch = tableUrlMatch(url);
  if (tableMatch) {
    const id = tableMatch[1];
    if (method === 'delete') {
      await updateCollectionSnapshots(TABLES_URL, ['tables', 'data'], (existing) =>
        removeById(existing, id)
      );
      return;
    }
    const patch = asRecord(responseBody) ?? asRecord(requestBody) ?? {};
    await updateCollectionSnapshots(TABLES_URL, ['tables', 'data'], (existing) =>
      existing.map((entry) =>
        sameId(entry, id)
          ? { ...(asRecord(entry) ?? {}), ...patch, id: idValue(entry, id) }
          : entry
      )
    );
    return;
  }

  const orderStatus = orderStatusForMutation(url, requestBody);
  if (orderStatus) {
    const orderId = orderIdFromMutationUrl(url);
    if (!orderId) return;
    await updateCollectionSnapshots(ORDERS_URL, ['orders', 'data'], (existing) =>
      existing.map((entry) =>
        sameId(entry, orderId)
          ? { ...(asRecord(entry) ?? {}), id: idValue(entry, orderId), status: orderStatus }
          : entry
      )
    );
    await patchCachedDetail(`${ORDERS_URL}/${orderId}`, { status: orderStatus });
    return;
  }

  const collectionUrl = collectionUrlForMutation(url);
  if (!collectionUrl) return;

  if (method === 'post') {
    const entry = asRecord(responseBody) ?? asRecord(requestBody);
    if (!entry) return;
    await updateCollectionSnapshots(collectionUrl, COLLECTION_KEYS[collectionUrl], (existing) =>
      upsertById(existing, entry)
    );
    return;
  }

  const itemId = finalPathId(url);
  if (!itemId) return;

  if (method === 'delete') {
    await updateCollectionSnapshots(collectionUrl, COLLECTION_KEYS[collectionUrl], (existing) =>
      removeById(existing, itemId)
    );
    return;
  }

  const patch = asRecord(responseBody) ?? asRecord(requestBody);
  if (!patch) return;
  await updateCollectionSnapshots(collectionUrl, COLLECTION_KEYS[collectionUrl], (existing) =>
    existing.map((entry) =>
      sameId(entry, itemId)
        ? { ...(asRecord(entry) ?? {}), ...patch, id: idValue(entry, itemId) }
        : entry
    )
  );
}

async function updateCollectionSnapshots(
  collectionUrl: string,
  listKeys: string[],
  update: (existing: unknown[]) => unknown[]
): Promise<void> {
  const snapshots = (await db.apiSnapshots.where('url').equals(collectionUrl).toArray()).filter(
    (snapshot) => snapshot.method === 'get'
  );
  if (snapshots.length === 0) {
    const responseBody = buildCollectionResponse(listKeys, update([]));
    await cacheResponse('get', collectionUrl, undefined, undefined, responseBody);
    return;
  }

  await Promise.all(
    snapshots.map((snapshot) =>
      db.apiSnapshots.put({
        ...snapshot,
        responseBody: updateResponseList(snapshot.responseBody, listKeys, update),
        updatedAt: Date.now(),
      })
    )
  );
}

async function patchCachedDetail(url: string, patch: RecordLike): Promise<void> {
  const snapshots = (await db.apiSnapshots.where('url').equals(url).toArray()).filter(
    (snapshot) => snapshot.method === 'get'
  );
  await Promise.all(
    snapshots.map((snapshot) => {
      const current = asRecord(snapshot.responseBody);
      if (!current) return Promise.resolve();
      return db.apiSnapshots.put({
        ...snapshot,
        responseBody: { ...current, ...patch },
        updatedAt: Date.now(),
      });
    })
  );
}

function updateResponseList(
  responseBody: unknown,
  listKeys: string[],
  update: (existing: unknown[]) => unknown[]
): unknown {
  if (Array.isArray(responseBody)) return update(responseBody);

  const root = asRecord(responseBody);
  if (!root) return buildCollectionResponse(listKeys, update([]));

  for (const key of listKeys) {
    if (Array.isArray(root[key])) {
      return { ...root, [key]: update(root[key] as unknown[]) };
    }
  }

  const data = asRecord(root.data);
  if (data) {
    for (const key of listKeys) {
      if (Array.isArray(data[key])) {
        return { ...root, data: { ...data, [key]: update(data[key] as unknown[]) } };
      }
    }
    if (Array.isArray(data.data)) {
      return { ...root, data: { ...data, data: update(data.data) } };
    }
  }

  if (Array.isArray(root.data)) {
    return { ...root, data: update(root.data) };
  }

  return { ...root, [listKeys[0] ?? 'data']: update([]) };
}

function buildCollectionResponse(listKeys: string[], items: unknown[]): unknown {
  return { [listKeys[0] ?? 'data']: items };
}

function buildOptimisticResponse(
  method: ApiMutationMethod,
  url: string,
  requestBody: unknown
): unknown {
  if (isOrdersCollectionUrl(url) && method === 'post') {
    return buildOptimisticOrder(requestBody);
  }
  if (isTablesCollectionUrl(url) && method === 'post') {
    return buildOptimisticTable(requestBody);
  }

  const id = positiveTempId();
  return {
    ...(asRecord(requestBody) ?? {}),
    id,
    offline: true,
    pendingSync: true,
  };
}

function buildOptimisticOrder(requestBody: unknown): RecordLike {
  const body = asRecord(requestBody) ?? {};
  const id = positiveTempId();
  const createdAt = new Date().toISOString();
  const items = Array.isArray(body.items) ? body.items : [];
  const normalizedItems = items.map((item, index) => {
    const raw = asRecord(item) ?? {};
    return {
      id: `${id}-${index + 1}`,
      name: text(raw.name, 'Item'),
      price: nonNegativeNumber(raw.unit_price ?? raw.unitPrice ?? raw.price),
      unitPrice: nonNegativeNumber(raw.unit_price ?? raw.unitPrice ?? raw.price),
      quantity: positiveNumber(raw.quantity, 1),
      notes: text(raw.notes, ''),
    };
  });
  const subtotal = normalizedItems.reduce(
    (sum, item) => sum + Number(item.price) * Number(item.quantity),
    0
  );

  return {
    ...body,
    id,
    orderId: id,
    order_id: id,
    number: `LOCAL-${String(id).slice(-6)}`,
    tableId: body.table_id ?? body.tableId,
    table_id: body.table_id ?? body.tableId,
    status: 'pending',
    items: normalizedItems,
    subtotal,
    tax: 0,
    tip: 0,
    discount: 0,
    total: subtotal,
    createdAt,
    created_at: createdAt,
    offline: true,
    pendingSync: true,
  };
}

function buildOptimisticTable(requestBody: unknown): RecordLike {
  const body = asRecord(requestBody) ?? {};
  const id = positiveTempId();
  const label = text(body.label ?? body.name ?? body.number, String(id));

  return {
    ...body,
    id,
    label,
    name: label,
    number: label,
    seats: positiveNumber(body.seats ?? body.capacity, 4),
    status: text(body.status, 'available'),
    offline: true,
    pendingSync: true,
  };
}

function mergeEntity(base: RecordLike, responseBody: unknown, pendingSync: boolean): RecordLike {
  const response = asRecord(responseBody) ?? {};
  const merged: RecordLike = { ...base, ...response };
  const resolvedId = idOf(response) ?? idOf(base);

  if (resolvedId) {
    merged.id = resolvedId;
    if ('orderId' in merged || 'order_id' in merged) {
      merged.orderId = resolvedId;
      merged.order_id = resolvedId;
    }
  }

  if (pendingSync) {
    merged.offline = true;
    merged.pendingSync = true;
  } else {
    delete merged.offline;
    delete merged.pendingSync;
  }

  return merged;
}

function orderStatusForMutation(url: string, requestBody: unknown): string | null {
  if (/^\/pos\/orders\/[^/]+\/status$/.test(url) || /^\/api\/pos\/orders\/v1\/orders\/[^/]+\/status$/.test(url)) {
    return text(asRecord(requestBody)?.status, 'pending');
  }
  if (/^\/pos\/orders\/[^/]+\/payment$/.test(url)) return 'paid';
  if (/^\/api\/pos\/orders\/payments\/v1\/orders\/[^/]+\/payment$/.test(url)) return 'paid';
  if (/^\/pos\/orders\/[^/]+\/void$/.test(url)) return 'voided';
  if (/^\/api\/pos\/orders\/v1\/orders\/[^/]+\/void$/.test(url)) return 'voided';
  if (/^\/pos\/orders\/[^/]+\/refund$/.test(url)) return 'refunded';
  if (/^\/api\/pos\/orders\/refunds\/v1\/orders\/[^/]+\/refund$/.test(url)) return 'refunded';
  if (/^\/pos\/orders\/[^/]+\/hold$/.test(url)) return 'held';
  if (/^\/api\/pos\/orders\/holds\/v1\/orders\/[^/]+\/hold$/.test(url)) return 'held';
  if (/^\/pos\/orders\/[^/]+\/send-to-kitchen$/.test(url)) return 'sent';
  return null;
}

function orderIdFromMutationUrl(url: string): string | null {
  return (
    url.match(/^\/pos\/orders\/([^/]+)/)?.[1] ??
    url.match(/^\/api\/pos\/orders\/(?:v1|payments\/v1|refunds\/v1|holds\/v1|discounts\/v1|splits\/v1)\/orders\/([^/]+)/)?.[1] ??
    null
  );
}

function isOrdersCollectionUrl(url: string): boolean {
  return url === ORDERS_URL || url === '/pos/orders';
}

function isTablesCollectionUrl(url: string): boolean {
  return url === TABLES_URL || url === '/pos/tables';
}

function tableStatusUrlMatch(url: string): RegExpMatchArray | null {
  return (
    url.match(/^\/pos\/tables\/([^/]+)\/status$/) ??
    url.match(/^\/api\/pos\/table\/v1\/tables\/([^/]+)$/)
  );
}

function tableUrlMatch(url: string): RegExpMatchArray | null {
  return url.match(/^\/pos\/tables\/([^/]+)$/) ?? url.match(/^\/api\/pos\/table\/v1\/tables\/([^/]+)$/);
}

function collectionUrlForMutation(url: string): string | null {
  for (const collectionUrl of Object.keys(COLLECTION_KEYS)) {
    if (collectionUrl === HELD_ORDERS_URL || collectionUrl === '/pos/orders/held') continue;
    if (url === collectionUrl || url.startsWith(`${collectionUrl}/`)) {
      return collectionUrl;
    }
  }
  return null;
}

function finalPathId(url: string): string | null {
  const parts = url.split('/').filter(Boolean);
  const last = parts[parts.length - 1];
  return last && /^\d+$/.test(last) ? last : null;
}

function upsertById(existing: unknown[], item: unknown): unknown[] {
  const itemId = idOf(item);
  if (!itemId) return [item, ...existing];

  const next = [...existing];
  const index = next.findIndex((entry) => sameId(entry, itemId));
  if (index === -1) {
    next.unshift(item);
  } else {
    next[index] = { ...(asRecord(next[index]) ?? {}), ...(asRecord(item) ?? {}) };
  }
  return next;
}

function removeById(existing: unknown[], id: unknown): unknown[] {
  return existing.filter((entry) => !sameId(entry, id));
}

function sameId(entry: unknown, id: unknown): boolean {
  const entryId = idOf(entry);
  if (entryId == null || id == null) return false;
  return String(entryId) === String(id);
}

function idOf(value: unknown): string | number | null {
  const record = asRecord(value);
  if (!record) return null;
  const id = record.id ?? record.orderId ?? record.order_id;
  if (typeof id === 'number' && Number.isFinite(id)) return id;
  const trimmed = String(id ?? '').trim();
  return trimmed || null;
}

function idValue(entry: unknown, fallback: unknown): unknown {
  return idOf(entry) ?? fallback;
}

function isQueueableMutation(url: string): boolean {
  if (!url.startsWith('/pos/') && !url.startsWith('/api/pos/')) return false;
  return !NON_QUEUEABLE_MUTATION_PATTERNS.some((pattern) => pattern.test(url));
}

function isRecoverableNetworkError(error: ApiError): boolean {
  return (
    error.status === 0 ||
    error.code === 'NETWORK_ERROR' ||
    error.code === 'TIMEOUT' ||
    !isOnline()
  );
}

function isOnline(): boolean {
  return typeof navigator === 'undefined' ? true : navigator.onLine;
}

function toApiError(error: unknown): ApiError {
  if (isApiError(error)) return error;
  return {
    status: 0,
    code: 'UNKNOWN',
    message: error instanceof Error ? error.message : 'An unexpected error occurred.',
    cause: error,
  };
}

function isApiError(value: unknown): value is ApiError {
  const error = asRecord(value);
  return (
    !!error &&
    typeof error.status === 'number' &&
    typeof error.code === 'string' &&
    typeof error.message === 'string'
  );
}

function networkApiError(message: string): ApiError {
  return {
    status: 0,
    code: 'NETWORK_ERROR',
    message,
  };
}

function requestKey(
  method: ApiRequestMethod,
  url: string,
  params?: unknown,
  requestBody?: unknown
): string {
  return stableStringify({
    method,
    url,
    params: params ?? null,
    requestBody: method === 'get' ? null : (requestBody ?? null),
  });
}

function stableStringify(value: unknown): string {
  if (Array.isArray(value)) {
    return `[${value.map(stableStringify).join(',')}]`;
  }
  const record = asRecord(value);
  if (!record) return JSON.stringify(value) ?? 'undefined';
  return `{${Object.keys(record)
    .sort()
    .map((key) => `${JSON.stringify(key)}:${stableStringify(record[key])}`)
    .join(',')}}`;
}

function storageSafe(value: unknown): unknown {
  if (value === undefined) return undefined;
  const serialized = JSON.stringify(value);
  return serialized === undefined ? undefined : JSON.parse(serialized);
}

function configFromParams(params: unknown): AxiosRequestConfig | undefined {
  return params === undefined ? undefined : { params };
}

async function logApiRequest(input: {
  method: ApiRequestMethod;
  url: string;
  params?: unknown;
  requestBody?: unknown;
  responseBody?: unknown;
  status: ApiRequestStatus;
  error?: string;
}): Promise<void> {
  const now = Date.now();
  await db.apiRequestLog.put({
    id: `req_${now}_${Math.random().toString(36).slice(2, 8)}`,
    ...input,
    createdAt: now,
    updatedAt: now,
  });
}

function dispatchLocalFirstEvent(name: string): void {
  if (typeof window === 'undefined') return;
  window.dispatchEvent(new CustomEvent(name));
}

function positiveTempId(): number {
  return Date.now() + Math.floor(Math.random() * 1000);
}

function asRecord(value: unknown): RecordLike | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as RecordLike)
    : null;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return 0;
  return numeric;
}

function positiveNumber(value: unknown, fallback: number): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return fallback;
  return numeric;
}

function text(value: unknown, fallback: string): string {
  const trimmed = String(value ?? '').trim();
  return trimmed || fallback;
}
