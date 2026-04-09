import React from 'react';
import POSLayout from '../../layouts/POSLayout';
import ProductGrid from './components/ProductGrid';
import OrderPanel from './components/OrderPanel';
import AiAssistant from './AiAssistant';
import styles from './POS.module.scss';

// Placeholder products — will be replaced by API data
const DEMO_CATEGORIES = ['Burgers', 'Sides', 'Drinks', 'Desserts'];

const DEMO_PRODUCTS = [
  { id: '1', name: 'Classic Burger', price: 12.99, category: 'Burgers' },
  { id: '2', name: 'Cheese Burger', price: 14.99, category: 'Burgers' },
  { id: '3', name: 'Veggie Burger', price: 11.99, category: 'Burgers' },
  { id: '4', name: 'BBQ Burger', price: 15.99, category: 'Burgers' },
  { id: '5', name: 'French Fries', price: 4.99, category: 'Sides' },
  { id: '6', name: 'Onion Rings', price: 5.99, category: 'Sides' },
  { id: '7', name: 'Coleslaw', price: 3.99, category: 'Sides' },
  { id: '8', name: 'Coca-Cola', price: 2.99, category: 'Drinks' },
  { id: '9', name: 'Lemonade', price: 3.49, category: 'Drinks' },
  { id: '10', name: 'Milkshake', price: 5.99, category: 'Drinks' },
  { id: '11', name: 'Brownie', price: 4.49, category: 'Desserts' },
  { id: '12', name: 'Ice Cream', price: 3.99, category: 'Desserts' },
];

const POS: React.FC = () => {
  return (
    <POSLayout>
      <div className={styles.pos}>
        <ProductGrid products={DEMO_PRODUCTS} categories={DEMO_CATEGORIES} />
        <OrderPanel />
      </div>
      <AiAssistant />
    </POSLayout>
  );
};

export default POS;
