/**
 * ERP application-wide constants.
 * Centralises magic strings and numbers used across features.
 */

// ─── Pagination ───────────────────────────────────────────────────────────────

export const DEFAULT_PAGE_SIZE = 50;
export const PAGE_SIZE_OPTIONS = [10, 25, 50, 100] as const;

// ─── Stock ────────────────────────────────────────────────────────────────────

export const DEFAULT_LOW_STOCK_THRESHOLD = 10;
export const DEFAULT_REORDER_SAFETY_FACTOR = 0.5;
export const DEFAULT_LEAD_TIME_DAYS = 7;

// ─── Price lists ──────────────────────────────────────────────────────────────

export const PRICE_LIST_TIERS = ['Retail', 'Wholesale', 'VIP', 'Custom'] as const;
export type PriceListTier = (typeof PRICE_LIST_TIERS)[number];

// ─── Purchase order statuses ──────────────────────────────────────────────────

export const PO_STATUSES = ['Draft', 'Sent', 'Received', 'Pending Approval', 'Approved'] as const;
export type POStatus = (typeof PO_STATUSES)[number];

// ─── Stock transfer statuses ──────────────────────────────────────────────────

export const TRANSFER_STATUSES = ['Draft', 'In Transit', 'Received', 'Cancelled'] as const;
export type TransferStatus = (typeof TRANSFER_STATUSES)[number];

// ─── Stock take statuses ──────────────────────────────────────────────────────

export const STOCK_TAKE_STATUSES = ['Draft', 'In Progress', 'Completed'] as const;
export type StockTakeStatus = (typeof STOCK_TAKE_STATUSES)[number];

// ─── Supplier statuses ────────────────────────────────────────────────────────

export const SUPPLIER_STATUSES = ['Active', 'Inactive'] as const;
export type SupplierStatus = (typeof SUPPLIER_STATUSES)[number];

// ─── Valuation methods ────────────────────────────────────────────────────────

export const VALUATION_METHODS = ['FIFO', 'LIFO', 'WAC'] as const;
export type ValuationMethod = (typeof VALUATION_METHODS)[number];

// ─── Activity types ───────────────────────────────────────────────────────────

export const ACTIVITY_TYPES = ['product', 'order', 'transfer', 'adjustment', 'supplier'] as const;
export type ActivityType = (typeof ACTIVITY_TYPES)[number];

// ─── Confidence levels ────────────────────────────────────────────────────────

export const CONFIDENCE_LEVELS = ['high', 'medium', 'low'] as const;
export type ConfidenceLevel = (typeof CONFIDENCE_LEVELS)[number];

// ─── Query stale times (ms) ───────────────────────────────────────────────────

export const STALE_TIME_PRODUCTS = 30_000;
export const STALE_TIME_SUPPLIERS = 60_000;
export const STALE_TIME_DASHBOARD = 15_000;
export const STALE_TIME_INVENTORY = 30_000;
