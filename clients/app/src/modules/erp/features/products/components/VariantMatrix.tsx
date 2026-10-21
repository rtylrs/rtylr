import React, { useMemo, useState } from 'react';
import { Tag, Plus, X, Trash2, Check, Wand2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { toast } from 'sonner';
import {
  parseVariantAttributes,
  useVariants,
  useCreateVariant,
  useUpdateVariant,
  useDeleteVariant,
  type Variant,
} from '../useVariants';

const ACCENT = 'var(--erp-accent)';

const inputStyle: React.CSSProperties = {
  padding: '6px 10px',
  borderRadius: '8px',
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: '13px',
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

interface OptionGroup {
  name: string;
  values: string[];
}

interface DraftVariant {
  attributes: Record<string, string>;
  name: string;
  sku: string;
  price: string;
  stock: string;
}

// Stable signature of an attribute set so we can dedupe combinations against
// variants that already exist (order-independent).
function attrSignature(attrs: Record<string, string>): string {
  return Object.keys(attrs)
    .sort()
    .map((k) => `${k}=${attrs[k]}`)
    .join('|');
}

function cartesian(groups: OptionGroup[]): Record<string, string>[] {
  const valid = groups.filter((g) => g.name.trim() && g.values.length > 0);
  if (valid.length === 0) return [];
  return valid.reduce<Record<string, string>[]>(
    (acc, g) => acc.flatMap((combo) => g.values.map((v) => ({ ...combo, [g.name.trim()]: v }))),
    [{}]
  );
}

function skuSuffix(attrs: Record<string, string>): string {
  return Object.values(attrs).join('-').toUpperCase().replace(/\s+/g, '');
}

interface VariantRowProps {
  variant: Variant;
}

const VariantRow: React.FC<VariantRowProps> = ({ variant }) => {
  const { t } = useTranslation();
  const updateMut = useUpdateVariant();
  const deleteMut = useDeleteVariant();
  const [sku, setSku] = useState(variant.sku ?? '');
  const [price, setPrice] = useState(String(variant.salePrice ?? 0));
  const [stock, setStock] = useState(String(variant.stockQuantity ?? 0));
  const [confirming, setConfirming] = useState(false);
  const attributes = useMemo(
    () => parseVariantAttributes(variant.attributes ?? ''),
    [variant.attributes]
  );

  const dirty =
    sku !== variant.sku ||
    price !== String(variant.salePrice) ||
    stock !== String(variant.stockQuantity);

  const handleSave = (): void => {
    updateMut.mutate(
      {
        id: variant.id,
        body: {
          productId: variant.productId,
          name: variant.name,
          sku: sku.trim(),
          barcode: variant.barcode,
          costPrice: variant.costPrice,
          salePrice: parseFloat(price) || 0,
          stockQuantity: parseFloat(stock) || 0,
          attributes: variant.attributes,
          isActive: variant.isActive,
        },
      },
      { onSuccess: () => toast.success(t('erp.products.variants.saved')) }
    );
  };

  const handleDelete = (): void => {
    deleteMut.mutate(variant.id ?? '', {
      onSuccess: () => toast.success(t('erp.products.variants.deleted')),
    });
  };

  return (
    <div
      className="grid items-center gap-2 px-3 py-2"
      style={{
        gridTemplateColumns: '1.6fr 1.4fr 0.9fr 0.9fr auto',
        borderTop: '1px solid var(--hub-border)',
      }}
    >
      <div className="min-w-0">
        <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
          {variant.name}
        </p>
        <div className="flex gap-1 flex-wrap mt-0.5">
          {Object.entries(attributes).map(([k, val]) => (
            <span
              key={k}
              className="text-[10px] px-1.5 py-0.5 rounded"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              {k}: {val}
            </span>
          ))}
        </div>
      </div>
      <input
        value={sku}
        onChange={(e) => setSku(e.target.value)}
        aria-label={t('erp.products.variants.sku')}
        style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '16px' }}
      />
      <input
        type="number"
        min="0"
        step="0.01"
        value={price}
        onChange={(e) => setPrice(e.target.value)}
        aria-label={t('erp.products.variants.price')}
        style={inputStyle}
      />
      <input
        type="number"
        min="0"
        step="1"
        value={stock}
        onChange={(e) => setStock(e.target.value)}
        aria-label={t('erp.products.variants.stock')}
        style={inputStyle}
      />
      <div className="flex items-center gap-1.5 justify-end">
        {dirty && (
          <button
            type="button"
            onClick={handleSave}
            disabled={updateMut.isPending}
            aria-label={t('erp.products.variants.save')}
            className="h-7 px-2 rounded-lg text-[11px] font-semibold disabled:opacity-40 flex items-center gap-1"
            style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
          >
            <Check size={12} /> {t('erp.products.variants.save')}
          </button>
        )}
        {confirming ? (
          <>
            <button
              type="button"
              onClick={handleDelete}
              disabled={deleteMut.isPending}
              aria-label={t('erpCommon.delete')}
              className="h-7 px-2 rounded-lg text-[11px] font-semibold disabled:opacity-40"
              style={{
                background: 'color-mix(in srgb, var(--erp-danger) 15%, transparent)',
                color: 'var(--erp-danger)',
              }}
            >
              {t('erpCommon.delete')}
            </button>
            <button
              type="button"
              onClick={() => setConfirming(false)}
              aria-label={t('erpCommon.cancel')}
              className="h-7 w-7 rounded-lg flex items-center justify-center"
              style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
            >
              <X size={12} />
            </button>
          </>
        ) : (
          <button
            type="button"
            onClick={() => setConfirming(true)}
            aria-label={t('erpCommon.delete')}
            className="h-7 w-7 rounded-lg flex items-center justify-center"
            style={{
              color: 'var(--erp-danger)',
              border: '1px solid color-mix(in srgb, var(--erp-danger) 30%, transparent)',
            }}
          >
            <Trash2 size={12} />
          </button>
        )}
      </div>
    </div>
  );
};

interface VariantMatrixProps {
  productId: string;
  productSku: string;
  baseSalePrice: number;
}

const VariantMatrix: React.FC<VariantMatrixProps> = ({ productId, productSku, baseSalePrice }) => {
  const { t } = useTranslation();
  const { data: variants = [] } = useVariants(productId);
  const createMut = useCreateVariant();

  const [groups, setGroups] = useState<OptionGroup[]>([]);
  const [valueDraft, setValueDraft] = useState<Record<number, string>>({});
  const [drafts, setDrafts] = useState<DraftVariant[]>([]);

  const existingSignatures = useMemo(
    () => new Set(variants.map((v) => attrSignature(parseVariantAttributes(v.attributes ?? '')))),
    [variants]
  );

  const combos = useMemo(() => cartesian(groups), [groups]);
  const newCombos = useMemo(
    () => combos.filter((c) => !existingSignatures.has(attrSignature(c))),
    [combos, existingSignatures]
  );

  const addGroup = (): void => setGroups((g) => [...g, { name: '', values: [] }]);

  const setGroupName = (idx: number, name: string): void =>
    setGroups((g) => g.map((grp, i) => (i === idx ? { ...grp, name } : grp)));

  const removeGroup = (idx: number): void => setGroups((g) => g.filter((_, i) => i !== idx));

  const addValue = (idx: number): void => {
    const raw = (valueDraft[idx] ?? '').trim();
    if (!raw) return;
    // Allow comma-separated bulk entry.
    const incoming = raw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    setGroups((g) =>
      g.map((grp, i) =>
        i === idx ? { ...grp, values: Array.from(new Set([...grp.values, ...incoming])) } : grp
      )
    );
    setValueDraft((d) => ({ ...d, [idx]: '' }));
  };

  const removeValue = (idx: number, value: string): void =>
    setGroups((g) =>
      g.map((grp, i) =>
        i === idx ? { ...grp, values: grp.values.filter((v) => v !== value) } : grp
      )
    );

  const generate = (): void => {
    if (newCombos.length === 0) {
      toast.error(t('erp.products.variants.allExist'));
      return;
    }
    setDrafts(
      newCombos.map((attributes) => ({
        attributes,
        name: Object.values(attributes).join(' / '),
        sku: `${productSku}-${skuSuffix(attributes)}`,
        price: baseSalePrice ? String(baseSalePrice) : '',
        stock: '0',
      }))
    );
  };

  const updateDraft = (idx: number, patch: Partial<DraftVariant>): void =>
    setDrafts((d) => d.map((row, i) => (i === idx ? { ...row, ...patch } : row)));

  const removeDraft = (idx: number): void => setDrafts((d) => d.filter((_, i) => i !== idx));

  const createAll = async (): Promise<void> => {
    if (drafts.length === 0) return;
    const batch = drafts;
    let created = 0;
    try {
      for (const d of batch) {
        await createMut.mutateAsync({
          productId,
          name: d.name.trim() || Object.values(d.attributes).join(' / '),
          sku: d.sku.trim(),
          barcode: '',
          costPrice: 0,
          salePrice: parseFloat(d.price) || 0,
          stockQuantity: parseFloat(d.stock) || 0,
          attributes: JSON.stringify(d.attributes),
          isActive: true,
        });
        created += 1;
        // Drop each row as soon as it exists server-side so a retry after a
        // partial failure cannot create it again.
        setDrafts((rows) =>
          rows.filter((row) => attrSignature(row.attributes) !== attrSignature(d.attributes))
        );
      }
      toast.success(t('erp.products.variants.created', { count: created }));
      setDrafts([]);
      setGroups([]);
      setValueDraft({});
    } catch {
      toast.error(
        created > 0
          ? t('erp.products.variants.createPartial', { count: created, total: batch.length })
          : t('erp.products.variants.createFailed')
      );
    }
  };

  return (
    <div
      className="rounded-xl p-5 space-y-4"
      style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
    >
      <div>
        <h3
          className="text-sm font-semibold flex items-center gap-2"
          style={{ color: 'var(--hub-text)' }}
        >
          <Tag size={14} style={{ color: ACCENT }} /> {t('erp.products.variants.title')}
        </h3>
        <p className="text-[11px] mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
          {t('erp.products.variants.subtitle')}
        </p>
      </div>

      {/* Existing variants */}
      {variants.length > 0 && (
        <div
          className="rounded-lg overflow-hidden"
          style={{ border: '1px solid var(--hub-border)' }}
        >
          <div
            className="grid gap-2 px-3 py-2 text-[10px] font-semibold uppercase tracking-wider"
            style={{
              gridTemplateColumns: '1.6fr 1.4fr 0.9fr 0.9fr auto',
              color: 'var(--hub-text-dim)',
              background: 'var(--hub-hover)',
            }}
          >
            <span>{t('erp.products.variants.variant')}</span>
            <span>{t('erp.products.variants.sku')}</span>
            <span>{t('erp.products.variants.price')}</span>
            <span>{t('erp.products.variants.stock')}</span>
            <span />
          </div>
          {variants.map((v) => (
            <VariantRow key={v.id} variant={v} />
          ))}
        </div>
      )}

      {/* Option groups builder */}
      <div className="space-y-3">
        {groups.map((grp, idx) => (
          <div
            key={idx}
            className="rounded-lg p-3 space-y-2"
            style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2">
              <input
                value={grp.name}
                onChange={(e) => setGroupName(idx, e.target.value)}
                placeholder={t('erp.products.variants.optionName')}
                style={inputStyle}
              />
              <button
                type="button"
                onClick={() => removeGroup(idx)}
                aria-label={t('erpCommon.remove')}
                className="h-8 w-8 shrink-0 rounded-lg flex items-center justify-center"
                style={{
                  color: 'var(--erp-danger)',
                  border: '1px solid color-mix(in srgb, var(--erp-danger) 30%, transparent)',
                }}
              >
                <X size={14} />
              </button>
            </div>
            <div className="flex gap-1.5 flex-wrap items-center">
              {grp.values.map((v) => (
                <span
                  key={v}
                  className="text-[11px] px-2 py-1 rounded-lg flex items-center gap-1"
                  style={{ background: 'var(--hub-card)', color: 'var(--hub-text)' }}
                >
                  {v}
                  <button
                    type="button"
                    onClick={() => removeValue(idx, v)}
                    aria-label={t('erpCommon.remove')}
                    style={{ color: 'var(--hub-text-dim)' }}
                  >
                    <X size={11} />
                  </button>
                </span>
              ))}
              <input
                value={valueDraft[idx] ?? ''}
                onChange={(e) => setValueDraft((d) => ({ ...d, [idx]: e.target.value }))}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ',') {
                    e.preventDefault();
                    addValue(idx);
                  }
                }}
                onBlur={() => addValue(idx)}
                placeholder={t('erp.products.variants.addValuePlaceholder')}
                style={{ ...inputStyle, width: 'auto', flex: '1 1 140px', minWidth: '140px' }}
              />
            </div>
          </div>
        ))}

        <div className="flex items-center gap-2 flex-wrap">
          <button
            type="button"
            onClick={addGroup}
            className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            <Plus size={13} /> {t('erp.products.variants.addOption')}
          </button>
          {combos.length > 0 && (
            <button
              type="button"
              onClick={generate}
              className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-semibold"
              style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
            >
              <Wand2 size={13} />{' '}
              {t('erp.products.variants.generateBtn', { count: newCombos.length })}
            </button>
          )}
        </div>
      </div>

      {/* Generated draft table */}
      {drafts.length > 0 && (
        <div className="space-y-3">
          <p className="text-xs font-semibold" style={{ color: 'var(--hub-text-muted)' }}>
            {t('erp.products.variants.preview')}
          </p>
          <div
            className="rounded-lg overflow-hidden"
            style={{ border: '1px solid var(--hub-border)' }}
          >
            {drafts.map((d, idx) => (
              <div
                key={attrSignature(d.attributes)}
                className="grid items-center gap-2 px-3 py-2"
                style={{
                  gridTemplateColumns: '1.6fr 1.4fr 0.9fr 0.9fr auto',
                  borderTop: idx > 0 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                  {d.name}
                </span>
                <input
                  value={d.sku}
                  onChange={(e) => updateDraft(idx, { sku: e.target.value })}
                  aria-label={t('erp.products.variants.sku')}
                  style={{ ...inputStyle, fontFamily: 'monospace', fontSize: '16px' }}
                />
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={d.price}
                  onChange={(e) => updateDraft(idx, { price: e.target.value })}
                  aria-label={t('erp.products.variants.price')}
                  style={inputStyle}
                />
                <input
                  type="number"
                  min="0"
                  step="1"
                  value={d.stock}
                  onChange={(e) => updateDraft(idx, { stock: e.target.value })}
                  aria-label={t('erp.products.variants.stock')}
                  style={inputStyle}
                />
                <button
                  type="button"
                  onClick={() => removeDraft(idx)}
                  aria-label={t('erpCommon.remove')}
                  className="h-7 w-7 rounded-lg flex items-center justify-center justify-self-end"
                  style={{ color: 'var(--hub-text-dim)', border: '1px solid var(--hub-border)' }}
                >
                  <X size={12} />
                </button>
              </div>
            ))}
          </div>
          <button
            type="button"
            onClick={createAll}
            disabled={createMut.isPending}
            className="px-4 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
            style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
          >
            {createMut.isPending
              ? t('erp.products.variants.creating')
              : t('erp.products.variants.createAll', { count: drafts.length })}
          </button>
        </div>
      )}

      {variants.length === 0 && drafts.length === 0 && groups.length === 0 && (
        <EmptyState
          icon={Tag}
          title={t('erp.products.variants.empty')}
          action={{ label: t('erp.products.variants.addOption'), onClick: addGroup }}
          size="compact"
        />
      )}
    </div>
  );
};

export default VariantMatrix;
