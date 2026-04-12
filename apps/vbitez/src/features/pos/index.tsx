import React, { useEffect, useState } from 'react';
import POSLayout from '../../layouts/POSLayout';
import QuickAccessSidebar from '../../shared/components/QuickAccessSidebar';
import ProductGrid from './components/ProductGrid';
import OrderPanel from './components/OrderPanel';
import AiAssistant from './AiAssistant';
import api from '../../shared/lib/api';

interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  image?: string;
  available?: boolean;
}

interface RawMenuItem {
  id?: string | number;
  name?: string;
  category?: string;
  categoryName?: string;
  price?: number;
  salePrice?: number;
  image?: string;
  imageUrl?: string;
  available?: boolean;
  isActive?: boolean;
}

const POS: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<string[]>([]);

  useEffect(() => {
    async function loadMenu() {
      try {
        const res = await api.get('/pos/menu/items');
        const items = (res.data?.items ?? res.data ?? []) as RawMenuItem[];
        const cats = new Set<string>();
        const mapped: Product[] = [];

        for (const item of items) {
          const cat = item.category ?? item.categoryName ?? 'Other';
          cats.add(cat);
          mapped.push({
            id: String(item.id),
            name: item.name ?? '',
            price: item.price ?? item.salePrice ?? 0,
            category: cat,
            image: item.image ?? item.imageUrl,
            available: item.available !== false && item.isActive !== false,
          });
        }

        setProducts(mapped);
        setCategories(Array.from(cats));
      } catch {
        // ProductGrid handles empty state
      }
    }

    loadMenu();
  }, []);

  return (
    <POSLayout>
      <div className="flex h-full overflow-hidden">
        <QuickAccessSidebar />
        <ProductGrid products={products} categories={categories} />
        <OrderPanel />
      </div>
      <AiAssistant />
    </POSLayout>
  );
};

export default POS;
