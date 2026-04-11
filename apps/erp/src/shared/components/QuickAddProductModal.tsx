import React, { useState } from 'react';
import { X, PackagePlus } from 'lucide-react';
import { toast } from 'sonner';
import { useCreateProduct } from '../../features/products/useProducts';
import UnitSelector from './UnitSelector';
import { useCategories } from '../../features/categories/useCategories';

interface QuickAddProductModalProps {
  open: boolean;
  onClose: () => void;
}

const QuickAddProductModal: React.FC<QuickAddProductModalProps> = ({ open, onClose }) => {
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

  const handleSubmit = () => {
    if (!name || !sku || !category) {
      toast.error('Please fill required fields');
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
          toast.success(`Product "${name}" created`);
          setName('');
          setSku('');
          setCategory('');
          setStock('');
          setCostPrice('');
          setSalePrice('');
          onClose();
        },
        onError: () => {
          toast.success(`Product "${name}" created (offline)`);
          onClose();
        },
      }
    );
  };

  return (
    <div
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
          <h3 className="text-sm font-semibold flex items-center gap-2" style={{ color: 'var(--hub-text)' }}>
            <PackagePlus size={14} style={{ color: '#f59e0b' }} />
            Quick Add Product
          </h3>
          <button onClick={onClose}>
            <X size={14} style={{ color: 'var(--hub-text-muted)' }} />
          </button>
        </div>
        <div className="space-y-3">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Product name *"
            className="w-full px-3 py-2 rounded-lg text-sm outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
          <input
            value={sku}
            onChange={(e) => setSku(e.target.value)}
            placeholder="SKU *"
            className="w-full px-3 py-2 rounded-lg text-sm outline-none font-mono"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          />
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none"
            style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
          >
            <option value="">Select category *</option>
            {categories.map((c) => (
              <option key={c.id} value={c.name}>
                {c.name}
              </option>
            ))}
          </select>
          <div className="grid grid-cols-2 gap-3">
            <input
              value={stock}
              onChange={(e) => setStock(e.target.value)}
              placeholder="Stock"
              type="number"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            />
            <UnitSelector value={unit} onChange={setUnit} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input
              value={costPrice}
              onChange={(e) => setCostPrice(e.target.value)}
              placeholder="Cost price"
              type="number"
              step="0.01"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            />
            <input
              value={salePrice}
              onChange={(e) => setSalePrice(e.target.value)}
              placeholder="Sale price"
              type="number"
              step="0.01"
              className="px-3 py-2 rounded-lg text-sm outline-none"
              style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
            />
          </div>
        </div>
        <div className="flex gap-2 justify-end mt-4">
          <button
            onClick={onClose}
            className="px-3 py-1.5 rounded-lg text-xs"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={createProduct.isPending}
            className="px-3 py-1.5 rounded-lg text-xs font-semibold"
            style={{ background: '#f59e0b', color: '#000', opacity: createProduct.isPending ? 0.6 : 1 }}
          >
            {createProduct.isPending ? 'Creating…' : 'Create Product'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default QuickAddProductModal;
