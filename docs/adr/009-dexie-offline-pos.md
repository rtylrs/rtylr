# ADR-009: Dexie (IndexedDB) for POS Offline-First

**Status:** Accepted  
**Date:** 2026-03-10  
**Decision:** The POS app (vbitez) uses Dexie.js for local IndexedDB storage, enabling offline operation.

## Context

POS terminals must work when internet connectivity is unreliable. A cashier can't tell a customer "sorry, the WiFi is down" during checkout. Options:

1. **Always-online** - require internet, show error when offline
2. **Service Worker cache** - cache API responses for offline reads
3. **IndexedDB via Dexie** - full local database with sync capability

## Decision

Dexie.js wrapping IndexedDB. The POS app maintains a local database with tables for: `products`, `modifiers`, `combos`, `customers`, `offlineOrders`.

On startup, the app syncs product catalog from the server. During operation, orders are created locally. When online, offline orders sync to the backend.

## Rationale

- **Instant checkout**: Product lookups are local IndexedDB queries (<1ms), not API calls (100-500ms). The cashier experience is snappy regardless of network.
- **Offline resilience**: Orders created during an outage are stored in `offlineOrders` and synced when connectivity returns. No lost sales.
- **Barcode/SKU indexing**: Dexie supports compound indexes. Products are indexed by barcode and SKU for instant scanner lookups.
- **Dexie over raw IndexedDB**: Dexie provides a clean Promise-based API, schema versioning, and TypeScript types. Raw IndexedDB is callback-based and painful.

## Trade-offs

- **Data staleness**: Product prices/availability may change on the server while the POS is offline. Sync frequency determines staleness window.
- **Storage limits**: IndexedDB has browser-dependent limits (typically 50-80% of available disk). Not an issue for typical product catalogs (<100MB).
- **Conflict resolution**: If the same order is modified both locally and on the server, conflict resolution is needed. Currently: last-write-wins.

## Consequences

- `apps/vbitez/src/shared/db/index.ts` defines the Dexie schema
- Product sync runs on POS login and periodically in background
- The checkout engine (`cartStore.ts`) operates entirely on local data
- Tauri desktop app uses the same Dexie/IndexedDB via its webview
