import React, { useEffect, useState } from 'react';
import POSLayout from '../../layouts/POSLayout';
import QuickAccessSidebar from '../../components/QuickAccessSidebar';
import ProductGrid from './components/ProductGrid';
import OrderPanel from './components/OrderPanel';
import AiAssistant from './AiAssistant';
import { businessApi, menuApi } from '../../api';

interface Product {
  id: string;
  name: string;
  price: number;
  category: string;
  image?: string;
  available?: boolean;
}

const POS: React.FC = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<string[]>([]);

  useEffect(() => {
    async function loadMenu() {
      try {
        const businesses = await businessApi.getAll();
        const businessId = Array.isArray(businesses) ? businesses[0]?.id : businesses?.id;
        if (!businessId) return;

        const menu = await menuApi.getMenu({ id: businessId });
        const items: Product[] = [];
        const cats = new Set<string>();

        const sections = menu?.categories ?? menu?.sections ?? menu?.menus ?? [];
        for (const section of sections) {
          const catName = section.name ?? section.category ?? 'Other';
          cats.add(catName);
          const sectionItems = section.items ?? section.products ?? [];
          for (const item of sectionItems) {
            items.push({
              id: String(item.id),
              name: item.name,
              price: item.price ?? item.salePrice ?? 0,
              category: catName,
              image: item.image ?? item.imageUrl,
              available: item.available !== false,
            });
          }
        }

        if (items.length > 0) {
          setProducts(items);
          setCategories(Array.from(cats));
        }
      } catch {
        // keep empty — ProductGrid handles empty state
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
