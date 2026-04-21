import React, { useState } from 'react';
import { X, PackagePlus } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useCreateProduct } from '../../features/products/useProducts';
import UnitSelector from './UnitSelector';
import { useCategories } from '../../features/categories/useCategories';

interface QuickAddProductModalProps {
  open: boolean;
  onClose: () => void;
}

const QuickAddProductModal: React.FC<QuickAddProductModalProps> = ({
  open,
  onClose,
}): React.JSX.Element | null => {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [sku, setSku] = useState('');
  const [category, setCategory] = useState('');
  const [stock, setStock] = useState('');
  const [costPrice, setCostPrice] = useState('');
  const [salePrice, setSalePrice] = useState('');
  const [unit, setUnit] = useState('pcs');
  const { data: categories = [] } = useCategories();
  const createProduct = useCreateProduct();

  if (!open) return null;

  const handleSubmit = (): void => {
    if (!name || !sku || !category) {
      toast.error(t('erp.products.validation.requiredFields'));
      return;
    }
    createProduct.mutate(
      {
        name,
        sku,
        category,
        stock: parseInt(stock) || 0,
        lowStock: 10,
        costPrice: parseFloat(costPrice) || 0,
        salePrice: parseFloat(salePrice) || 0,
        unit,
      },
      {
        onSuccess: () => {
          toast.success(t('erp.products.toast.created', { name }));
          setName('');
          setSku('');
          setCategory('');
          setStock('');
          setCostPrice('');
          setSalePrice('');
          onClose();
        },
        onError: () => {
          // Surface a real failure toast instead of faking success — the
          // previous behaviour masked backend outages and lost user data.
          toast.error(t('erp.products.toast.createFailed'));
        },
      }
    );
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('erp.products.new')}
      data-testid="quick-add-product-modal"
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'rgba(0,0,0,0.6)' }}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-md rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <h3
            className="text-sm font-semibold flex items-center gap-2"
            style={{ color: 'var(--hub-text)' }}
          >
            <PackagePlus size={14} style={{ color: '#f59e0b' }} />
            {t('erp.products.new')}
          </h3>
          <button onClick={onClose} aria-label={t('common.close')}>
            <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
          </button>
        </div>
        <div className="space-y-3">
          <label className="sr-only" htmlFor="qap-name">
            {t('erp.products.fields.name')}
          </label>
          <input
            id="qap-name"
            data-testid="qap-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={`${t('erp.products.fields.name')} *`}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
          <label className="sr-only" htmlFor="qap-sku">
            {t('erp.products.fields.sku')}
          </label>
          <input
            id="qap-sku"
            data-testid="qap-sku"
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            placeholder={`${t('erp.products.fields.sku')} *`}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none font-mono"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
          <label className="sr-only" htmlFor="qap-category">
            {t('erp.products.fields.category')}
          </label>
          <select
            id="qap-category"
            data-testid="qap-category"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="">{t('erp.products.fields.selectCategory')} *</option>
            {categories.map((c) => (
              <option key={c.id} value={c.name}>
                {c.name}
              </option>
            ))}
          </select>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.stock')}
              value={stock}
              onChange={(e) => setStock(e.target.value)}
              placeholder={t('erp.products.fields.stock')}
              type="number"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <UnitSelector value={unit} onChange={setUnit} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.costPrice')}
              value={costPrice}
              onChange={(e) => setCostPrice(e.target.value)}
              placeholder={t('erp.products.fields.costPrice')}
              type="number"
              step="0.01"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              aria-label={t('erp.products.fields.salePrice')}
              value={salePrice}
              onChange={(e) => setSalePrice(e.target.value)}
              placeholder={t('erp.products.fields.salePrice')}
              type="number"
              step="0.01"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
          </div>
        </div>
        <div className="flex gap-2 justify-end mt-4">
          <button
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg text-xs"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            {t('common.cancel')}
          </button>
          <button
            data-testid="qap-submit"
            onClick={handleSubmit}
            disabled={createProduct.isPending}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{
              background: '#f59e0b',
              color: '#000',
              opacity: createProduct.isPending ? 0.6 : 1,
            }}
          >
            {createProduct.isPending ? t('erp.products.creating') : t('erp.products.create')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default QuickAddProductModal;
