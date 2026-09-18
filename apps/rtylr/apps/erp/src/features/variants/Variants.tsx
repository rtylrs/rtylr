import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Layers } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { parseVariantAttributes, useVariants, useCreateVariant } from '../products/useVariants';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { formatCurrency } from '../../shared/lib/format';
import StockBadge from '../../shared/components/StockBadge';
import { toast } from 'sonner';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  borderRadius: '8px',
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: '13px',
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

const Variants: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data: variants = [], isLoading } = useVariants();
  const createMut = useCreateVariant();
  const [showCreate, setShowCreate] = useState(false);
  const [newProductId, setNewProductId] = useState('');
  const [newName, setNewName] = useState('');
  const [newSku, setNewSku] = useState('');
  const [newPrice, setNewPrice] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!newProductId.trim()) {
      toast.error(t('erp.variants.productIdRequired'));
      return;
    }
    if (!newName.trim()) {
      toast.error(t('erp.variants.nameRequired'));
      return;
    }
    const price = parseFloat(newPrice) || 0;
    if (price < 0) {
      toast.error(t('erp.variants.priceNegative'));
      return;
    }
    try {
      await createMut.mutateAsync({
        productId: newProductId.trim(),
        name: newName.trim(),
        sku: newSku.trim(),
        barcode: '',
        costPrice: 0,
        salePrice: price,
        stockQuantity: 0,
        attributes: '{}',
        isActive: true,
      });
      toast.success(t('erp.variants.created'));
      setShowCreate(false);
      setNewProductId('');
      setNewName('');
      setNewSku('');
      setNewPrice('');
    } catch {
      toast.error(t('erp.variants.createFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/variants'].title))} />
      <PageHeader
        title={t(routeLabels['/variants'].title)}
        description={t('erp.variants.subtitle')}
        actions={
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
          >
            <Plus size={14} /> New Variant
          </button>
        }
      />

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
          style={{
            gridTemplateColumns: '2fr 1fr 1.5fr 0.8fr 0.8fr 0.8fr',
            color: 'var(--hub-text-dim)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span>{t('erp.products.variants.variant')}</span>
          <span>{t('erp.products.variants.sku')}</span>
          <span>{t('erp.variants.attributes')}</span>
          <span>{t('erp.products.variants.stock')}</span>
          <span>{t('common.price')}</span>
          <span>{t('common.status')}</span>
        </div>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={6} />)
        ) : variants.length === 0 ? (
          <EmptyState
            icon={Layers}
            title={t('erp.variants.empty.title')}
            description={t('erp.variants.empty.description')}
            action={{ label: t('erp.variants.addVariant'), onClick: () => setShowCreate(true) }}
          />
        ) : (
          variants.map((v, i) => {
            const attributes = parseVariantAttributes(v.attributes ?? '');
            return (
              <div
                key={v.id}
                className="grid px-4 py-3 items-center"
                style={{
                  gridTemplateColumns: '2fr 1fr 1.5fr 0.8fr 0.8fr 0.8fr',
                  borderTop: i > 0 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {v.name}
                </span>
                <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                  {v.sku}
                </span>
                <div className="flex gap-1 flex-wrap">
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
                <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {v.stockQuantity}
                </span>
                <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(v.salePrice ?? 0)}
                </span>
                <StockBadge stock={v.stockQuantity ?? 0} lowStock={10} />
              </div>
            );
          })
        )}
      </div>

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'var(--hub-overlay)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.variants.newVariant')}
            </h3>
            <input
              value={newProductId}
              onChange={(e) => setNewProductId(e.target.value)}
              placeholder={t('erp.variants.productIdPlaceholder')}
              style={inputStyle}
            />
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder={t('erp.variants.namePlaceholder')}
              style={inputStyle}
            />
            <input
              value={newSku}
              onChange={(e) => setNewSku(e.target.value)}
              placeholder={t('erp.variants.skuPlaceholder')}
              style={inputStyle}
            />
            <input
              type="number"
              min="0"
              step="0.01"
              value={newPrice}
              onChange={(e) => setNewPrice(e.target.value)}
              placeholder={t('common.price')}
              style={inputStyle}
            />
            <div className="flex gap-2 justify-end">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                onClick={handleCreate}
                disabled={createMut.isPending || !newProductId.trim() || !newName.trim()}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Variants;
