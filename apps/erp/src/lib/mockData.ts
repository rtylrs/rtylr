import {
  Product,
  Category,
  Brand,
  Warehouse,
  StockTransfer,
  StockTake,
  ProductVariant,
  PriceList,
  ExpiryItem,
  ReorderSuggestion,
  ActivityItem,
  Supplier,
  PurchaseOrder,
  LowStockItem,
} from '../services/erpApi';

export const mockProducts: Product[] = [
  { id: 1, name: 'Arabica Coffee Beans', sku: 'BEV-001', category: 'Beverages', brand: 'Mountain Roast', stock: 3, lowStock: 20, costPrice: 18.5, salePrice: 29.99, status: 'Low Stock', barcode: '8901234567890', unit: 'kg', description: 'Premium Arabica coffee beans from Ethiopia', tags: ['coffee', 'premium'], reorderPoint: 25, leadTimeDays: 7 },
  { id: 2, name: 'Whole Milk (2L)', sku: 'BEV-012', category: 'Beverages', brand: 'Dairy Fresh', stock: 8, lowStock: 50, costPrice: 2.1, salePrice: 3.5, status: 'Low Stock', barcode: '8901234567891', unit: 'pcs', description: 'Fresh whole milk 2 litre bottles', tags: ['dairy'], reorderPoint: 60, leadTimeDays: 2 },
  { id: 3, name: 'Sourdough Bread', sku: 'FOOD-034', category: 'Bakery', brand: 'Artisan Bakers', stock: 0, lowStock: 15, costPrice: 3.0, salePrice: 5.5, status: 'Out of Stock', barcode: '8901234567892', unit: 'pcs', description: 'Traditional sourdough loaf', tags: ['bread', 'fresh'], reorderPoint: 18, leadTimeDays: 1 },
  { id: 4, name: 'Extra Virgin Olive Oil (1L)', sku: 'FOOD-055', category: 'Pantry', brand: 'Olive Grove', stock: 42, lowStock: 10, costPrice: 7.8, salePrice: 14.99, status: 'In Stock', barcode: '8901234567893', unit: 'pcs', description: 'Cold pressed Italian olive oil', tags: ['oil', 'italian'], reorderPoint: 15, leadTimeDays: 10 },
  { id: 5, name: 'Sparkling Water (500ml)', sku: 'BEV-028', category: 'Beverages', brand: 'Crystal Springs', stock: 120, lowStock: 30, costPrice: 0.6, salePrice: 1.5, status: 'In Stock', barcode: '8901234567894', unit: 'pcs', description: 'Carbonated natural mineral water', tags: ['water'], reorderPoint: 40, leadTimeDays: 3 },
  { id: 6, name: 'Cheddar Cheese (500g)', sku: 'DAIRY-007', category: 'Dairy', brand: 'Farmhouse', stock: 18, lowStock: 15, costPrice: 5.2, salePrice: 9.99, status: 'In Stock', barcode: '8901234567895', unit: 'pcs', description: 'Aged cheddar cheese block', tags: ['cheese'], reorderPoint: 20, leadTimeDays: 4 },
  { id: 7, name: 'Organic Honey (500g)', sku: 'FOOD-089', category: 'Pantry', brand: 'BeeWild', stock: 35, lowStock: 12, costPrice: 6.4, salePrice: 12.5, status: 'In Stock', barcode: '8901234567896', unit: 'pcs', description: 'Raw wildflower honey', tags: ['honey', 'organic'], reorderPoint: 18, leadTimeDays: 14 },
  { id: 8, name: 'Greek Yogurt (1kg)', sku: 'DAIRY-022', category: 'Dairy', brand: 'Dairy Fresh', stock: 4, lowStock: 20, costPrice: 3.1, salePrice: 6.0, status: 'Low Stock', barcode: '8901234567897', unit: 'pcs', description: 'Traditional strained Greek yogurt', tags: ['yogurt'], reorderPoint: 25, leadTimeDays: 2 },
  { id: 9, name: 'Dark Chocolate Bar (100g)', sku: 'SNACK-011', category: 'Snacks', brand: 'Cacao Co', stock: 88, lowStock: 25, costPrice: 1.8, salePrice: 4.5, status: 'In Stock', barcode: '8901234567898', unit: 'pcs', description: '70% dark chocolate bar', tags: ['chocolate'], reorderPoint: 30, leadTimeDays: 7 },
  { id: 10, name: 'Basmati Rice (5kg)', sku: 'PANTRY-045', category: 'Pantry', brand: 'Gold Grain', stock: 26, lowStock: 8, costPrice: 9.5, salePrice: 18.99, status: 'In Stock', barcode: '8901234567899', unit: 'bags', description: 'Premium long grain basmati rice', tags: ['rice', 'grain'], reorderPoint: 12, leadTimeDays: 10 },
  { id: 11, name: 'Green Tea Bags (50ct)', sku: 'BEV-067', category: 'Beverages', brand: 'Zen Garden', stock: 54, lowStock: 20, costPrice: 3.2, salePrice: 7.5, status: 'In Stock', barcode: '8901234567900', unit: 'boxes', description: 'Japanese sencha green tea', tags: ['tea'], reorderPoint: 25, leadTimeDays: 14 },
  { id: 12, name: 'Almond Butter (340g)', sku: 'PANTRY-112', category: 'Pantry', brand: 'Nut House', stock: 19, lowStock: 10, costPrice: 6.0, salePrice: 11.99, status: 'In Stock', barcode: '8901234567901', unit: 'pcs', description: 'Natural smooth almond butter', tags: ['nut', 'butter'], reorderPoint: 15, leadTimeDays: 7 },
];

export const mockCategories: Category[] = [
  { id: 1, name: 'Beverages', description: 'Drinks and liquid refreshments', parentId: null, productCount: 48, sortOrder: 1 },
  { id: 2, name: 'Dairy', description: 'Milk, cheese, yogurt and related products', parentId: null, productCount: 32, sortOrder: 2 },
  { id: 3, name: 'Bakery', description: 'Fresh baked goods', parentId: null, productCount: 24, sortOrder: 3 },
  { id: 4, name: 'Pantry', description: 'Dry goods and shelf staples', parentId: null, productCount: 67, sortOrder: 4 },
  { id: 5, name: 'Snacks', description: 'Chips, chocolates and quick bites', parentId: null, productCount: 41, sortOrder: 5 },
  { id: 6, name: 'Frozen', description: 'Frozen foods and ice cream', parentId: null, productCount: 19, sortOrder: 6 },
  { id: 7, name: 'Produce', description: 'Fresh fruits and vegetables', parentId: null, productCount: 52, sortOrder: 7 },
  { id: 8, name: 'Household', description: 'Cleaning and home supplies', parentId: null, productCount: 15, sortOrder: 8 },
];

export const mockBrands: Brand[] = [
  { id: 1, name: 'Mountain Roast', description: 'Premium coffee roasters since 1998', productCount: 12 },
  { id: 2, name: 'Dairy Fresh', description: 'Farm-to-table dairy products', productCount: 24 },
  { id: 3, name: 'Artisan Bakers', description: 'Traditional baking methods', productCount: 18 },
  { id: 4, name: 'Olive Grove', description: 'Mediterranean olive oils', productCount: 8 },
  { id: 5, name: 'Crystal Springs', description: 'Natural mineral waters', productCount: 6 },
  { id: 6, name: 'Farmhouse', description: 'Artisan cheese producers', productCount: 15 },
  { id: 7, name: 'BeeWild', description: 'Wild honey and bee products', productCount: 5 },
  { id: 8, name: 'Cacao Co', description: 'Fine chocolate makers', productCount: 22 },
];

export const mockWarehouses: Warehouse[] = [
  { id: 1, name: 'Main Distribution Center', address: '1200 Industrial Ave, Austin TX', manager: 'Sarah Chen', capacity: 10000, used: 7200, isDefault: true },
  { id: 2, name: 'North Regional Hub', address: '45 Logistics Way, Dallas TX', manager: 'Marcus Johnson', capacity: 6500, used: 4100, isDefault: false },
  { id: 3, name: 'West Coast Depot', address: '880 Harbor Blvd, Long Beach CA', manager: 'Elena Rodriguez', capacity: 8000, used: 5400, isDefault: false },
  { id: 4, name: 'East Storage Facility', address: '77 Commerce Rd, Newark NJ', manager: 'David Kim', capacity: 4500, used: 2800, isDefault: false },
];

export const mockStockTransfers: StockTransfer[] = [
  { id: 1001, fromWarehouseId: 1, toWarehouseId: 2, fromName: 'Main Distribution Center', toName: 'North Regional Hub', items: 45, status: 'In Transit', createdAt: '2026-04-07' },
  { id: 1002, fromWarehouseId: 1, toWarehouseId: 3, fromName: 'Main Distribution Center', toName: 'West Coast Depot', items: 120, status: 'Received', createdAt: '2026-04-05', completedAt: '2026-04-08' },
  { id: 1003, fromWarehouseId: 2, toWarehouseId: 4, fromName: 'North Regional Hub', toName: 'East Storage Facility', items: 32, status: 'Draft', createdAt: '2026-04-09' },
  { id: 1004, fromWarehouseId: 3, toWarehouseId: 1, fromName: 'West Coast Depot', toName: 'Main Distribution Center', items: 18, status: 'Cancelled', createdAt: '2026-04-03' },
];

export const mockStockTakes: StockTake[] = [
  { id: 501, name: 'Q2 2026 Monthly Count', warehouseId: 1, warehouseName: 'Main Distribution Center', status: 'In Progress', itemsCounted: 142, totalItems: 248, variance: 4, createdAt: '2026-04-08' },
  { id: 502, name: 'Cold Storage Audit', warehouseId: 2, warehouseName: 'North Regional Hub', status: 'Completed', itemsCounted: 89, totalItems: 89, variance: 2, createdAt: '2026-04-02' },
  { id: 503, name: 'West Coast Spot Check', warehouseId: 3, warehouseName: 'West Coast Depot', status: 'Draft', itemsCounted: 0, totalItems: 156, variance: 0, createdAt: '2026-04-09' },
];

export const mockVariants: ProductVariant[] = [
  { id: 201, productId: 1, name: 'Arabica 250g', sku: 'BEV-001-250', attributes: { size: '250g' }, stock: 24, price: 9.99 },
  { id: 202, productId: 1, name: 'Arabica 500g', sku: 'BEV-001-500', attributes: { size: '500g' }, stock: 18, price: 18.50 },
  { id: 203, productId: 1, name: 'Arabica 1kg', sku: 'BEV-001-1000', attributes: { size: '1kg' }, stock: 12, price: 34.99 },
  { id: 204, productId: 9, name: 'Dark Chocolate Mint', sku: 'SNACK-011-MINT', attributes: { flavor: 'Mint' }, stock: 45, price: 4.50 },
  { id: 205, productId: 9, name: 'Dark Chocolate Orange', sku: 'SNACK-011-ORANGE', attributes: { flavor: 'Orange' }, stock: 38, price: 4.50 },
];

export const mockPriceLists: PriceList[] = [
  { id: 1, name: 'Standard Retail', description: 'Default retail pricing for walk-in customers', tier: 'Retail', multiplier: 1.0, activeCount: 248 },
  { id: 2, name: 'Wholesale Tier A', description: 'Pricing for bulk wholesale buyers', tier: 'Wholesale', multiplier: 0.75, activeCount: 180 },
  { id: 3, name: 'VIP Members', description: 'Exclusive pricing for VIP program members', tier: 'VIP', multiplier: 0.85, activeCount: 210 },
  { id: 4, name: 'Restaurant Partners', description: 'Custom pricing for restaurant partners', tier: 'Custom', multiplier: 0.70, activeCount: 95 },
];

export const mockExpiryItems: ExpiryItem[] = [
  { productId: 2, name: 'Whole Milk (2L)', sku: 'BEV-012', batchId: 'B-2604', expiryDate: '2026-04-14', quantity: 8, warehouseName: 'Main Distribution Center' },
  { productId: 3, name: 'Sourdough Bread', sku: 'FOOD-034', batchId: 'B-2605', expiryDate: '2026-04-11', quantity: 12, warehouseName: 'Main Distribution Center' },
  { productId: 8, name: 'Greek Yogurt (1kg)', sku: 'DAIRY-022', batchId: 'B-2602', expiryDate: '2026-04-18', quantity: 20, warehouseName: 'North Regional Hub' },
  { productId: 6, name: 'Cheddar Cheese (500g)', sku: 'DAIRY-007', batchId: 'B-2598', expiryDate: '2026-05-20', quantity: 18, warehouseName: 'Main Distribution Center' },
];

export const mockReorderSuggestions: ReorderSuggestion[] = [
  { productId: 1, name: 'Arabica Coffee Beans', sku: 'BEV-001', currentStock: 3, avgDailySales: 2.4, suggestedQuantity: 40, leadTimeDays: 7, confidence: 'high' },
  { productId: 2, name: 'Whole Milk (2L)', sku: 'BEV-012', currentStock: 8, avgDailySales: 12.0, suggestedQuantity: 80, leadTimeDays: 2, confidence: 'high' },
  { productId: 3, name: 'Sourdough Bread', sku: 'FOOD-034', currentStock: 0, avgDailySales: 6.2, suggestedQuantity: 30, leadTimeDays: 1, confidence: 'high' },
  { productId: 8, name: 'Greek Yogurt (1kg)', sku: 'DAIRY-022', currentStock: 4, avgDailySales: 3.1, suggestedQuantity: 30, leadTimeDays: 2, confidence: 'medium' },
];

export const mockActivity: ActivityItem[] = [
  { id: 1, type: 'order', title: 'PO-91 created', description: 'New purchase order for Artisan Bakers Inc.', user: 'Sarah Chen', timestamp: '2026-04-09T10:30:00Z' },
  { id: 2, type: 'adjustment', title: 'Stock adjusted', description: 'Arabica Coffee Beans +5 (recount)', user: 'Marcus Johnson', timestamp: '2026-04-09T09:45:00Z' },
  { id: 3, type: 'product', title: 'New product added', description: 'Almond Butter (340g) added to catalog', user: 'Elena Rodriguez', timestamp: '2026-04-09T08:20:00Z' },
  { id: 4, type: 'transfer', title: 'Transfer completed', description: 'Stock transfer 1002 received', user: 'David Kim', timestamp: '2026-04-08T16:15:00Z' },
  { id: 5, type: 'supplier', title: 'Supplier updated', description: 'Global Beverages Ltd. contact info changed', user: 'Sarah Chen', timestamp: '2026-04-08T14:00:00Z' },
  { id: 6, type: 'order', title: 'PO-90 sent', description: 'PO sent to Global Beverages Ltd.', user: 'Sarah Chen', timestamp: '2026-04-08T11:00:00Z' },
  { id: 7, type: 'adjustment', title: 'Variance recorded', description: 'Cold Storage Audit variance: 2 items', user: 'Marcus Johnson', timestamp: '2026-04-07T17:30:00Z' },
  { id: 8, type: 'product', title: 'Price updated', description: 'Extra Virgin Olive Oil price: $12.99 → $14.99', user: 'Elena Rodriguez', timestamp: '2026-04-07T15:45:00Z' },
];

export const mockSuppliers: Supplier[] = [
  { id: 1, name: 'Fresh Farms Co.', contactName: 'Anna Reed', email: 'anna@freshfarms.co', phone: '+1 512 555 0101', productCount: 45, categories: ['Produce', 'Dairy'], status: 'Active', onTimeDeliveryPct: 94, fulfillmentAccuracy: 98, leadTimeDays: 3 },
  { id: 2, name: 'Global Beverages Ltd.', contactName: 'Carlos Ruiz', email: 'carlos@globalbev.com', phone: '+1 214 555 0102', productCount: 32, categories: ['Beverages'], status: 'Active', onTimeDeliveryPct: 88, fulfillmentAccuracy: 95, leadTimeDays: 5 },
  { id: 3, name: 'Artisan Bakers Inc.', contactName: 'Mei Lin', email: 'mei@artisanbakers.com', phone: '+1 310 555 0103', productCount: 18, categories: ['Bakery'], status: 'Active', onTimeDeliveryPct: 99, fulfillmentAccuracy: 100, leadTimeDays: 1 },
  { id: 4, name: 'Pantry Wholesale', contactName: 'Jorge Ortiz', email: 'jorge@pantrywhole.com', phone: '+1 602 555 0104', productCount: 67, categories: ['Pantry', 'Snacks'], status: 'Active', onTimeDeliveryPct: 82, fulfillmentAccuracy: 92, leadTimeDays: 7 },
];

export const mockPurchaseOrders: PurchaseOrder[] = [
  { id: 89, supplier: 'Fresh Farms Co.', supplierId: 1, date: '2026-04-07', expectedDate: '2026-04-10', items: 12, total: '$1,240.00', status: 'Received' },
  { id: 90, supplier: 'Global Beverages Ltd.', supplierId: 2, date: '2026-04-08', expectedDate: '2026-04-12', items: 20, total: '$3,560.00', status: 'Sent' },
  { id: 91, supplier: 'Artisan Bakers Inc.', supplierId: 3, date: '2026-04-09', expectedDate: '2026-04-13', items: 4, total: '$890.00', status: 'Draft' },
  { id: 92, supplier: 'Pantry Wholesale', supplierId: 4, date: '2026-04-09', expectedDate: '2026-04-16', items: 28, total: '$5,220.00', status: 'Pending Approval' },
];

export const mockLowStockItems: LowStockItem[] = [
  { name: 'Arabica Coffee Beans', sku: 'BEV-001', stock: 3, minStock: 20, unit: 'kg' },
  { name: 'Whole Milk (2L)', sku: 'BEV-012', stock: 8, minStock: 50, unit: 'units' },
  { name: 'Sourdough Bread', sku: 'FOOD-034', stock: 0, minStock: 15, unit: 'loaves' },
  { name: 'Greek Yogurt (1kg)', sku: 'DAIRY-022', stock: 4, minStock: 20, unit: 'pcs' },
];
