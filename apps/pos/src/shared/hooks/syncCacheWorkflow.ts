import type { CachedCombo, CachedModifier, CachedProduct } from '../services/db';

type RawRecord = Record<string, unknown>;

function asRecord(value: unknown): RawRecord | undefined {
  return value && typeof value === 'object' ? (value as RawRecord) : undefined;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return 0;
  return numeric;
}

function booleanValue(value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback;
}

function text(value: unknown, fallback = ''): string {
  return typeof value === 'string' ? value.trim() || fallback : fallback;
}

function optionalText(value: unknown): string | undefined {
  const trimmed = text(value);
  return trimmed || undefined;
}

export function responseArray(response: unknown, keys: string[]): unknown[] {
  if (Array.isArray(response)) return response;
  const record = asRecord(response);
  if (!record) return [];

  for (const key of keys) {
    const value = record[key];
    if (Array.isArray(value)) return value;
  }

  return [];
}

export function normalizeCachedProducts(raw: unknown[], syncedAt: number): CachedProduct[] {
  return raw.flatMap((value) => {
    const product = asRecord(value);
    const id = positiveInteger(product?.id);
    if (!product || id == null) return [];

    return [
      {
        id,
        name: text(product.name, 'Item'),
        sku: text(product.sku),
        barcode: optionalText(product.barcode),
        categoryId: positiveInteger(product.categoryId ?? product.category_id) ?? 0,
        categoryName: text(product.categoryName ?? product.category, 'Other'),
        price: nonNegativeNumber(product.price ?? product.salePrice ?? product.sale_price),
        costPrice: nonNegativeNumber(product.costPrice ?? product.cost_price),
        taxClass: text(product.taxClass ?? product.tax_class, 'standard'),
        imageUrl: optionalText(product.image ?? product.imageUrl ?? product.image_url),
        available:
          booleanValue(product.available, true) &&
          booleanValue(product.isActive ?? product.is_active, true),
        syncedAt,
      },
    ];
  });
}

export function normalizeCachedModifiers(raw: unknown[], syncedAt: number): CachedModifier[] {
  return raw.flatMap((value) => {
    const record = asRecord(value);
    if (!record) return [];

    if (Array.isArray(record.modifiers)) {
      const groupName = text(record.name ?? record.groupName ?? record.group_name, 'Modifiers');
      return normalizeCachedModifiers(
        record.modifiers.map((modifier) => ({ ...(asRecord(modifier) ?? {}), groupName })),
        syncedAt,
      );
    }

    const id = positiveInteger(record.id);
    if (id == null) return [];

    return [
      {
        id,
        name: text(record.name, 'Option'),
        groupName: text(record.groupName ?? record.group_name ?? record.group, 'Modifiers'),
        priceDelta: nonNegativeNumber(record.priceDelta ?? record.price_delta),
        isRequired: booleanValue(record.isRequired ?? record.is_required ?? record.required, false),
        syncedAt,
      },
    ];
  });
}

export function normalizeCachedCombos(raw: unknown[], syncedAt: number): CachedCombo[] {
  return raw.flatMap((value) => {
    const combo = asRecord(value);
    const id = positiveInteger(combo?.id);
    if (!combo || id == null) return [];

    return [
      {
        id,
        name: text(combo.name, 'Combo'),
        description: text(combo.description),
        price: nonNegativeNumber(combo.price),
        items: normalizeComboItems(Array.isArray(combo.items) ? combo.items : []),
        isActive: booleanValue(combo.isActive ?? combo.is_active, true),
        syncedAt,
      },
    ];
  });
}

function normalizeComboItems(raw: unknown[]): CachedCombo['items'] {
  return raw.flatMap((value) => {
    const item = asRecord(value);
    const productId = positiveInteger(item?.productId ?? item?.product_id ?? item?.itemId);
    if (!item || productId == null) return [];

    return [
      {
        productId,
        name: text(item.name, 'Item'),
        quantity: positiveInteger(item.quantity ?? item.qty) ?? 1,
      },
    ];
  });
}
