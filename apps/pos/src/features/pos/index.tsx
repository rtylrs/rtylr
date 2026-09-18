import React, { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import POSLayout from '../../layouts/POSLayout';
import QuickAccessSidebar from '../../shared/components/QuickAccessSidebar';
import ProductGrid from './components/ProductGrid';
import OrderPanel from './components/OrderPanel';
import { useCartStore } from '../../shared/stores/cartStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { useMenuItems } from '../menu-items/useMenuItems';

const POS: React.FC = () => {
  const { items: products, categories } = useMenuItems();
  const [searchParams] = useSearchParams();
  const setTableNumber = useCartStore((s) => s.setTableNumber);
  const setTable = useCartStore((s) => s.setTable);
  const tablesEnabled = useSettingsStore((s) => s.features.tables);
  const tableParam = searchParams.get('table');
  const tableNumber = tablesEnabled ? (tableParam ?? undefined) : undefined;

  useEffect(() => {
    setTableNumber(tableNumber ?? null);
    if (!tablesEnabled) {
      setTable(null, null);
    }
  }, [tableNumber, setTable, setTableNumber, tablesEnabled]);

  return (
    <POSLayout>
      <div className="flex h-full overflow-hidden">
        <QuickAccessSidebar />
        <ProductGrid products={products} categories={categories} />
        <OrderPanel tableNumber={tableNumber} />
      </div>
    </POSLayout>
  );
};

export default POS;
