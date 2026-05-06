import React, { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import POSLayout from '../../layouts/POSLayout';
import QuickAccessSidebar from '../../shared/components/QuickAccessSidebar';
import ProductGrid from './components/ProductGrid';
import OrderPanel from './components/OrderPanel';
import AiAssistant from './AiAssistant';
import { useCartStore } from '../../shared/stores/cartStore';
import { useMenuItems } from '../menu-items/useMenuItems';

const POS: React.FC = () => {
  const { items: products, categories } = useMenuItems();
  const [searchParams] = useSearchParams();
  const setTableNumber = useCartStore((s) => s.setTableNumber);
  const tableNumber = searchParams.get('table') ?? undefined;

  useEffect(() => {
    if (tableNumber) {
      setTableNumber(tableNumber);
    }
  }, [tableNumber, setTableNumber]);

  return (
    <POSLayout>
      <div className="flex h-full overflow-hidden">
        <QuickAccessSidebar />
        <ProductGrid products={products} categories={categories} />
        <OrderPanel tableNumber={tableNumber} />
      </div>
      <AiAssistant />
    </POSLayout>
  );
};

export default POS;
