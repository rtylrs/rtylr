import React, { useState } from 'react';
import { motion } from 'motion/react';
import { useNavigate } from '@tanstack/react-router';
import { Printer, Check } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { fetchAllProducts, type Product } from '../products/useProducts';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import BarcodeSvg from '../../shared/components/BarcodeSvg';
import { toast } from 'sonner';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

const Barcodes: React.FC = () => {
  const { t } = useTranslation('erp');
  const navigate = useNavigate();
  const { data: products = [], isLoading } = useQuery<Product[]>({
    queryKey: ['products', 'barcodes'],
    queryFn: () => fetchAllProducts(),
  });
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [copies, setCopies] = useState(1);

  const toggle = (id: string) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelected(next);
  };

  const handlePrint = () => {
    if (selected.size === 0) {
      toast.error(t('erp.barcodes.selectFirstToast'));
      return;
    }
    toast.success(t('erp.barcodes.printing', { count: selected.size * copies }));
    setTimeout(() => window.print(), 100);
  };

  const selectedProducts = products.filter((p) => selected.has(p.id ?? ''));

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/barcodes'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/barcodes'].title)}
        description={t('erp.barcodes.subtitle')}
        actions={
          <div className="flex items-center gap-2">
            <div
              className="flex items-center gap-1.5 px-2 py-1.5 rounded-lg"
              style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
            >
              <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.barcodes.copies')}
              </span>
              <input
                type="number"
                min="1"
                value={copies}
                onChange={(e) => setCopies(Math.max(1, parseInt(e.target.value) || 1))}
                className="w-12 bg-transparent text-xs font-semibold outline-none"
                style={{ color: 'var(--hub-text)' }}
              />
            </div>
            <button
              type="button"
              onClick={handlePrint}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
            >
              <Printer size={14} /> Print {selected.size > 0 ? `(${selected.size * copies})` : ''}
            </button>
          </div>
        }
      />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div
          className="lg:col-span-1 rounded-xl p-3 max-h-[calc(100vh-220px)] overflow-y-auto"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <p className="text-xs font-semibold mb-2 px-2" style={{ color: 'var(--hub-text-muted)' }}>
            {t('erp.barcodes.products')}
          </p>
          {isLoading
            ? Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-10 mb-1" />)
            : products.map((p) => (
                <button
                  type="button"
                  key={p.id}
                  onClick={() => toggle(p.id ?? '')}
                  className="w-full flex items-center gap-2 px-2 py-2 rounded-lg text-left"
                  style={{
                    background: selected.has(p.id ?? '') ? 'var(--hub-active)' : 'transparent',
                  }}
                >
                  <div
                    className="h-4 w-4 rounded flex items-center justify-center"
                    style={{
                      background: selected.has(p.id ?? '') ? AMBER : 'var(--hub-hover)',
                      border: '1px solid var(--hub-border)',
                    }}
                  >
                    {selected.has(p.id ?? '') && (
                      <Check size={10} style={{ color: 'var(--dash-text)' }} />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p
                      className="text-xs font-medium truncate"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {p.name}
                    </p>
                    <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                      {p.sku}
                    </p>
                  </div>
                </button>
              ))}
        </div>

        <div className="lg:col-span-2">
          {selectedProducts.length === 0 ? (
            <div
              className="h-full rounded-xl flex items-center justify-center"
              style={{ background: 'var(--hub-card)', border: '1px dashed var(--hub-border)' }}
            >
              <EmptyState
                icon={Printer}
                title={t('erp.barcodes.previewEmpty.title')}
                description={t('erp.barcodes.previewEmpty.description')}
                size="compact"
                action={
                  products.length > 0
                    ? {
                        label: t('erp.barcodes.selectFirst'),
                        onClick: () => toggle(products[0]?.id ?? ''),
                      }
                    : {
                        label: t('erp.barcodes.goToProducts'),
                        onClick: () => navigate({ to: '/erp/products' }),
                      }
                }
              />
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              {selectedProducts.flatMap((p) =>
                Array.from({ length: copies }).map((_, i) => (
                  <div
                    key={`${p.id}-${i}`}
                    className="rounded-lg p-3 bg-white"
                    style={{ border: '1px solid var(--hub-border)' }}
                  >
                    <p className="text-[11px] font-bold text-black mb-1 truncate">{p.name}</p>
                    <BarcodeSvg value={p.barcode || p.sku || ''} width={200} height={50} />
                    <p className="text-[10px] text-black/60 mt-1">
                      ${(p.salePrice ?? 0).toFixed(2)}
                    </p>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default Barcodes;
