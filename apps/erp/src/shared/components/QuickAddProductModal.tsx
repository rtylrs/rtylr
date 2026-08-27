import React, { useEffect, useId, useRef, useState } from 'react';
import { X, PackagePlus, Plus, Upload } from 'lucide-react';
import { toast } from 'sonner';
import { uploadAsset } from '@voxire/api';
import { cdnImageUrl } from '@voxire/config';
import { useTranslation } from 'react-i18next';
import {
  useCreateProduct,
  useProductAttributes,
  useProductFull,
  useUpdateProduct,
  type CreateProductInput,
  type Product,
} from '../../features/products/useProducts';
import UnitSelector from './UnitSelector';
import { useDialogFocus } from '../hooks/useDialogFocus';
import { useCategories } from '../../features/categories/useCategories';
import { useSuppliers } from '../../features/suppliers/useSuppliers';
import { useBrands } from '../../features/brands/useBrands';

interface QuickAddProductModalProps {
  open: boolean;
  onClose: () => void;
  /**
   * When supplied, the modal switches to edit mode: fields are seeded from
   * the product and submit triggers the update mutation instead of create.
   */
  editing?: Product | null;
}

const inputClass = 'w-full px-3 py-2 rounded-lg text-sm outline-none';
const inputStyle: React.CSSProperties = {
  background: 'var(--hub-input-bg)',
  border: '1px solid var(--hub-border)',
  color: 'var(--hub-text)',
};

const SectionLabel: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div
    className="text-[11px] font-semibold uppercase tracking-wide mt-1"
    style={{ color: 'var(--hub-text-muted)' }}
  >
    {children}
  </div>
);

const QuickAddProductModal: React.FC<QuickAddProductModalProps> = ({
  open,
  onClose,
  editing,
}): React.JSX.Element | null => {
  const { t } = useTranslation();
  const isEdit = !!editing;

  const [name, setName] = useState('');
  const [sku, setSku] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [supplierId, setSupplierId] = useState('');
  const [barcode, setBarcode] = useState('');
  const [imageUrl, setImageUrl] = useState('');
  const [description, setDescription] = useState('');
  const [unit, setUnit] = useState('pcs');
  const [costPrice, setCostPrice] = useState('');
  const [salePrice, setSalePrice] = useState('');
  const [stock, setStock] = useState('');
  const [lowStock, setLowStock] = useState('');
  const [reorderPoint, setReorderPoint] = useState('');
  const [trackInventory, setTrackInventory] = useState(true);
  const [active, setActive] = useState(true);
  const [brandId, setBrandId] = useState('');
  const [compareAtPrice, setCompareAtPrice] = useState('');
  const [taxRate, setTaxRate] = useState('');
  const [currency, setCurrency] = useState('USD');
  const [weight, setWeight] = useState('');
  const [length, setLength] = useState('');
  const [width, setWidth] = useState('');
  const [height, setHeight] = useState('');
  const [countryOfOrigin, setCountryOfOrigin] = useState('');
  const [hsCode, setHsCode] = useState('');
  const [leadTimeDays, setLeadTimeDays] = useState('');
  const [allowBackorder, setAllowBackorder] = useState(false);
  const [attributes, setAttributes] = useState<{ name: string; value: string }[]>([]);
  const [uploadingImage, setUploadingImage] = useState(false);
  const imageInputRef = useRef<HTMLInputElement>(null);

  // Prefill custom attributes when editing an existing product.
  const { data: attributesData } = useProductAttributes(editing?.id);
  useEffect(() => {
    if (attributesData) {
      setAttributes(
        attributesData.map((attribute) => ({
          name: attribute.name ?? '',
          value: attribute.value ?? '',
        }))
      );
    }
  }, [attributesData]);

  // Authoritative full-product fetch on edit: the list payload lacks the PIM
  // fields, so without this an edit-save would zero compare-at/tax/weight/etc.
  const { data: fullProduct } = useProductFull(editing?.id);
  useEffect(() => {
    const p = fullProduct;
    if (!p) return;
    setName(p.name ?? '');
    setSku(p.sku ?? '');
    setCategoryId(p.categoryId !== '0' ? (p.categoryId ?? '') : '');
    setSupplierId(p.supplierId !== '0' ? (p.supplierId ?? '') : '');
    setBrandId(p.brandId !== '0' ? (p.brandId ?? '') : '');
    setBarcode(p.barcode ?? '');
    setImageUrl(p.imageUrl ?? editing?.imageUrl ?? '');
    setDescription(p.description ?? '');
    setUnit(p.unit || 'pcs');
    setCostPrice(p.costPrice ? String(p.costPrice) : '');
    setSalePrice(p.salePrice ? String(p.salePrice) : '');
    setStock(p.stockQuantity ? String(p.stockQuantity) : '');
    setLowStock(p.lowStockThreshold ? String(p.lowStockThreshold) : '');
    setReorderPoint(p.reorderQuantity ? String(p.reorderQuantity) : '');
    setTrackInventory(p.trackInventory ?? false);
    setActive(p.isActive ?? false);
    setCompareAtPrice(p.compareAtPrice ? String(p.compareAtPrice) : '');
    setTaxRate(p.taxRate ? String(p.taxRate) : '');
    setCurrency(p.currency || 'USD');
    setWeight(p.weight ? String(p.weight) : '');
    setLength(p.length ? String(p.length) : '');
    setWidth(p.width ? String(p.width) : '');
    setHeight(p.height ? String(p.height) : '');
    setCountryOfOrigin(p.countryOfOrigin ?? '');
    setHsCode(p.hsCode ?? '');
    setLeadTimeDays(p.leadTimeDays ? String(p.leadTimeDays) : '');
    setAllowBackorder(p.allowBackorder ?? false);
  }, [fullProduct]);

  const { data: categories = [] } = useCategories();
  const { data: supplierData } = useSuppliers();
  const suppliers = supplierData?.suppliers ?? [];
  const { data: brands = [] } = useBrands();
  const createProduct = useCreateProduct();
  const updateProduct = useUpdateProduct();
  const titleId = useId();
  const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose });

  // Re-seed local state when a different product is loaded for editing.
  useEffect(() => {
    if (editing) {
      setName(editing.name ?? '');
      setSku(editing.sku ?? '');
      setCategoryId(
        editing.categoryId !== undefined && editing.categoryId !== null ? editing.categoryId : ''
      );
      setSupplierId(
        editing.supplierId !== undefined && editing.supplierId !== null ? editing.supplierId : ''
      );
      setBarcode(editing.barcode ?? '');
      setImageUrl((editing as Product & { imageUrl?: string }).imageUrl ?? '');
      setDescription(editing.description ?? '');
      setUnit(editing.unit ?? 'pcs');
      setCostPrice(String(editing.costPrice ?? ''));
      setSalePrice(String(editing.salePrice ?? ''));
      setStock(editing.stockQuantity != null ? String(editing.stockQuantity) : '');
      setLowStock(editing.lowStockThreshold != null ? String(editing.lowStockThreshold) : '');
      setReorderPoint(editing.reorderQuantity != null ? String(editing.reorderQuantity) : '');
    }
  }, [editing]);

  if (!open) return null;

  const pending = createProduct.isPending || updateProduct.isPending || uploadingImage;
  const cost = parseFloat(costPrice);
  const sale = parseFloat(salePrice);
  const margin =
    !isNaN(cost) && !isNaN(sale) && sale > 0 ? Math.round(((sale - cost) / sale) * 100) : null;

  const resetAndClose = (): void => {
    if (!isEdit) {
      setName('');
      setSku('');
      setCategoryId('');
      setSupplierId('');
      setBarcode('');
      setImageUrl('');
      setDescription('');
      setUnit('pcs');
      setCostPrice('');
      setSalePrice('');
      setStock('');
      setLowStock('');
      setReorderPoint('');
      setTrackInventory(true);
      setActive(true);
      setBrandId('');
      setCompareAtPrice('');
      setTaxRate('');
      setCurrency('USD');
      setWeight('');
      setLength('');
      setWidth('');
      setHeight('');
      setCountryOfOrigin('');
      setHsCode('');
      setLeadTimeDays('');
      setAllowBackorder(false);
      setAttributes([]);
    }
    onClose();
  };

  const addAttribute = (): void => setAttributes((a) => [...a, { name: '', value: '' }]);
  const updateAttribute = (i: number, key: 'name' | 'value', val: string): void =>
    setAttributes((a) => a.map((x, idx) => (idx === i ? { ...x, [key]: val } : x)));
  const removeAttribute = (i: number): void =>
    setAttributes((a) => a.filter((_, idx) => idx !== i));

  const handleImageUpload = async (file?: File | null): Promise<void> => {
    if (!file) return;
    setUploadingImage(true);
    try {
      const uploaded = await uploadAsset(file);
      setImageUrl(uploaded.url);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploadingImage(false);
      if (imageInputRef.current) imageInputRef.current.value = '';
    }
  };

  const handleSubmit = (): void => {
    if (!name || !sku || !categoryId || !supplierId || !barcode || !description || !unit) {
      toast.error(t('erp.products.validation.requiredFields'));
      return;
    }
    if (!isNaN(cost) && cost < 0) {
      toast.error(t('erp.products.validation.negativeCost'));
      return;
    }
    if (!isNaN(sale) && sale < 0) {
      toast.error(t('erp.products.validation.negativePrice'));
      return;
    }

    const payload: CreateProductInput = {
      name,
      sku,
      categoryId,
      supplierId,
      barcode,
      description,
      unit,
      costPrice: isNaN(cost) ? 0 : cost,
      salePrice: isNaN(sale) ? 0 : sale,
      stockQuantity: stock ? Number(stock) : 0,
      lowStockThreshold: lowStock ? Number(lowStock) : 0,
      reorderQuantity: reorderPoint ? Number(reorderPoint) : 0,
      isActive: active,
      trackInventory,
      brandId: brandId || '0',
      compareAtPrice: compareAtPrice ? Number(compareAtPrice) : 0,
      taxRate: taxRate ? Number(taxRate) : 0,
      currency,
      weight: weight ? Number(weight) : 0,
      weightUnit: 'kg',
      length: length ? Number(length) : 0,
      width: width ? Number(width) : 0,
      height: height ? Number(height) : 0,
      dimensionUnit: 'cm',
      countryOfOrigin,
      hsCode,
      allowBackorder,
      leadTimeDays: leadTimeDays ? Number(leadTimeDays) : 0,
      imageUrl,
    };

    const onSuccess = (): void => {
      toast.success(t('erp.products.toast.created', { name }));
      resetAndClose();
    };
    const onError = (): void => {
      toast.error(t('erp.products.toast.createFailed'));
    };

    if (isEdit && editing) {
      updateProduct.mutate({ id: editing.id, body: payload }, { onSuccess, onError });
      return;
    }
    createProduct.mutate(payload, { onSuccess, onError });
  };

  return (
    <div
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      data-testid="quick-add-product-modal"
      className="fixed inset-0 z-50 flex items-center justify-center p-6"
      style={{ background: 'var(--hub-overlay-heavy)' }}
      tabIndex={-1}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-lg rounded-xl p-5 max-h-[88vh] overflow-y-auto"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center justify-between mb-4">
          <h3
            id={titleId}
            className="text-sm font-semibold flex items-center gap-2"
            style={{ color: 'var(--hub-text)' }}
          >
            <PackagePlus size={14} style={{ color: 'var(--erp-accent)' }} />
            {t('erp.products.new')}
          </h3>
          <button type="button" onClick={onClose} aria-label={t('common.close')}>
            <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
          </button>
        </div>

        <div className="space-y-3">
          <SectionLabel>{t('erp.products.sections.basics')}</SectionLabel>
          <input
            data-testid="qap-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={`${t('erp.products.fields.name')} *`}
            className={inputClass}
            style={inputStyle}
          />
          <input
            data-testid="qap-sku"
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            placeholder={`${t('erp.products.fields.sku')} *`}
            className={`${inputClass} font-mono`}
            style={inputStyle}
          />
          <select
            data-testid="qap-category"
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className={inputClass}
            style={inputStyle}
          >
            <option value="">{t('erp.products.fields.selectCategory')} *</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
          <select
            aria-label={t('erp.products.fields.supplier')}
            value={supplierId}
            onChange={(e) => setSupplierId(e.target.value)}
            className={inputClass}
            style={inputStyle}
          >
            <option value="">{t('erp.products.fields.selectSupplier')}</option>
            {suppliers.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <select
            aria-label={t('erp.products.fields.brand')}
            value={brandId}
            onChange={(e) => setBrandId(e.target.value)}
            className={inputClass}
            style={inputStyle}
          >
            <option value="">{t('erp.products.fields.selectBrand')}</option>
            {brands.map((b) => (
              <option key={b.id} value={b.id}>
                {b.name}
              </option>
            ))}
          </select>
          <input
            aria-label={t('erp.products.fields.barcode')}
            value={barcode}
            onChange={(e) => setBarcode(e.target.value)}
            placeholder={t('erp.products.fields.barcode')}
            className={`${inputClass} font-mono`}
            style={inputStyle}
          />
          <textarea
            aria-label={t('erp.products.fields.description')}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('erp.products.fields.description')}
            rows={2}
            className={`${inputClass} resize-none`}
            style={inputStyle}
          />
          <div className="grid grid-cols-[1fr_auto] gap-2">
            <input
              aria-label="Product image URL"
              value={imageUrl}
              onChange={(e) => setImageUrl(e.target.value)}
              placeholder="Product image URL"
              className={inputClass}
              style={inputStyle}
            />
            <button
              type="button"
              onClick={() => imageInputRef.current?.click()}
              disabled={uploadingImage}
              className="px-3 py-2 rounded-lg text-sm font-semibold inline-flex items-center gap-2"
              style={{
                background: 'var(--hub-hover)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            >
              <Upload size={14} />
              {uploadingImage ? 'Uploading...' : 'Upload'}
            </button>
            <input
              ref={imageInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => {
                void handleImageUpload(e.target.files?.[0]);
              }}
            />
          </div>
          {imageUrl && (
            <img
              src={cdnImageUrl(imageUrl)}
              alt=""
              className="h-28 w-full rounded-lg object-contain"
              style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
            />
          )}

          <SectionLabel>{t('erp.products.sections.pricing')}</SectionLabel>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.costPrice')}
              value={costPrice}
              onChange={(e) => setCostPrice(e.target.value)}
              placeholder={t('erp.products.fields.costPrice')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.salePrice')}
              value={salePrice}
              onChange={(e) => setSalePrice(e.target.value)}
              placeholder={t('erp.products.fields.salePrice')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
          </div>
          {margin !== null && (
            <div className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              {t('erp.products.fields.margin')}:{' '}
              <span style={{ color: 'var(--erp-success)' }}>{margin}%</span>
            </div>
          )}
          <div className="grid grid-cols-3 gap-3">
            <input
              aria-label={t('erp.products.fields.compareAt')}
              value={compareAtPrice}
              onChange={(e) => setCompareAtPrice(e.target.value)}
              placeholder={t('erp.products.fields.compareAt')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.taxRate')}
              value={taxRate}
              onChange={(e) => setTaxRate(e.target.value)}
              placeholder={t('erp.products.fields.taxRate')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.currency')}
              value={currency}
              onChange={(e) => setCurrency(e.target.value.toUpperCase().slice(0, 3))}
              placeholder={t('erp.products.fields.currency')}
              maxLength={3}
              className={`${inputClass} uppercase`}
              style={inputStyle}
            />
          </div>

          <SectionLabel>{t('erp.products.sections.inventory')}</SectionLabel>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.stock')}
              value={stock}
              onChange={(e) => setStock(e.target.value)}
              placeholder={t('erp.products.fields.stock')}
              type="number"
              className={inputClass}
              style={inputStyle}
            />
            <UnitSelector value={unit} onChange={setUnit} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.lowStock')}
              value={lowStock}
              onChange={(e) => setLowStock(e.target.value)}
              placeholder={t('erp.products.fields.lowStock')}
              type="number"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.reorderPoint')}
              value={reorderPoint}
              onChange={(e) => setReorderPoint(e.target.value)}
              placeholder={t('erp.products.fields.reorderPoint')}
              type="number"
              className={inputClass}
              style={inputStyle}
            />
          </div>
          <div className="flex items-center justify-between gap-3 pt-1">
            <label
              className="flex items-center gap-2 text-xs cursor-pointer"
              style={{ color: 'var(--hub-text)' }}
            >
              <input
                type="checkbox"
                checked={trackInventory}
                onChange={(e) => setTrackInventory(e.target.checked)}
              />
              {t('erp.products.fields.trackInventory')}
            </label>
            <label
              className="flex items-center gap-2 text-xs cursor-pointer"
              style={{ color: 'var(--hub-text)' }}
            >
              <input
                type="checkbox"
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
              />
              {t('erp.products.fields.statusActive')}
            </label>
          </div>

          <SectionLabel>{t('erp.products.sections.logistics')}</SectionLabel>
          <input
            aria-label={t('erp.products.fields.weight')}
            value={weight}
            onChange={(e) => setWeight(e.target.value)}
            placeholder={t('erp.products.fields.weight')}
            type="number"
            step="0.001"
            className={inputClass}
            style={inputStyle}
          />
          <div className="grid grid-cols-3 gap-3">
            <input
              aria-label={t('erp.products.fields.length')}
              value={length}
              onChange={(e) => setLength(e.target.value)}
              placeholder={t('erp.products.fields.length')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.width')}
              value={width}
              onChange={(e) => setWidth(e.target.value)}
              placeholder={t('erp.products.fields.width')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.height')}
              value={height}
              onChange={(e) => setHeight(e.target.value)}
              placeholder={t('erp.products.fields.height')}
              type="number"
              step="0.01"
              className={inputClass}
              style={inputStyle}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input
              aria-label={t('erp.products.fields.countryOfOrigin')}
              value={countryOfOrigin}
              onChange={(e) => setCountryOfOrigin(e.target.value.toUpperCase().slice(0, 2))}
              placeholder={t('erp.products.fields.countryOfOrigin')}
              maxLength={2}
              className={`${inputClass} uppercase`}
              style={inputStyle}
            />
            <input
              aria-label={t('erp.products.fields.hsCode')}
              value={hsCode}
              onChange={(e) => setHsCode(e.target.value)}
              placeholder={t('erp.products.fields.hsCode')}
              className={`${inputClass} font-mono`}
              style={inputStyle}
            />
          </div>
          <div className="grid grid-cols-2 gap-3 items-center">
            <input
              aria-label={t('erp.products.fields.leadTime')}
              value={leadTimeDays}
              onChange={(e) => setLeadTimeDays(e.target.value)}
              placeholder={t('erp.products.fields.leadTime')}
              type="number"
              className={inputClass}
              style={inputStyle}
            />
            <label
              className="flex items-center gap-2 text-xs cursor-pointer"
              style={{ color: 'var(--hub-text)' }}
            >
              <input
                type="checkbox"
                checked={allowBackorder}
                onChange={(e) => setAllowBackorder(e.target.checked)}
              />
              {t('erp.products.fields.allowBackorder')}
            </label>
          </div>

          <SectionLabel>{t('erp.products.sections.attributes')}</SectionLabel>
          {attributes.map((a, i) => (
            <div key={i} className="flex gap-2 items-center">
              <input
                aria-label={t('erp.products.fields.attrName')}
                value={a.name}
                onChange={(e) => updateAttribute(i, 'name', e.target.value)}
                placeholder={t('erp.products.fields.attrName')}
                className={inputClass}
                style={inputStyle}
              />
              <input
                aria-label={t('erp.products.fields.attrValue')}
                value={a.value}
                onChange={(e) => updateAttribute(i, 'value', e.target.value)}
                placeholder={t('erp.products.fields.attrValue')}
                className={inputClass}
                style={inputStyle}
              />
              <button
                type="button"
                onClick={() => removeAttribute(i)}
                aria-label={t('common.remove')}
                className="shrink-0 rounded-lg p-1.5"
                style={{ background: 'var(--hub-hover)' }}
              >
                <X size={13} style={{ color: 'var(--hub-text-muted)' }} />
              </button>
            </div>
          ))}
          <button
            type="button"
            onClick={addAttribute}
            className="flex items-center gap-1.5 text-xs"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <Plus size={13} /> {t('erp.products.fields.addAttribute')}
          </button>
        </div>

        <div className="flex gap-2 justify-end mt-4">
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg text-xs"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            {t('common.cancel')}
          </button>
          <button
            type="button"
            data-testid="qap-submit"
            onClick={handleSubmit}
            disabled={pending}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{
              background: 'var(--erp-accent)',
              color: 'var(--erp-accent-foreground)',
              opacity: pending ? 0.6 : 1,
            }}
          >
            {pending
              ? t('erp.products.creating')
              : isEdit
                ? t('common.save')
                : t('erp.products.create')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default QuickAddProductModal;
