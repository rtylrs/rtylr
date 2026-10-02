export { listProducts, getProduct, getProductByBarcode, createProduct, updateProduct, deleteProduct } from './products.js';
export { listSuppliers, createSupplier, updateSupplier, deleteSupplier } from './suppliers.js';
export { listPurchaseOrders, getPurchaseOrder, createPurchaseOrder, approvePurchaseOrder } from './purchase-orders.js';
export { inventoryReport, listPriceLists, createPriceList, listStockTransfers, listStockTakes } from './inventory.js';

import { listProducts, getProduct, getProductByBarcode, createProduct, updateProduct, deleteProduct } from './products.js';
import { listSuppliers, createSupplier, updateSupplier, deleteSupplier } from './suppliers.js';
import { listPurchaseOrders, getPurchaseOrder, createPurchaseOrder, approvePurchaseOrder } from './purchase-orders.js';
import { inventoryReport, listPriceLists, createPriceList, listStockTransfers, listStockTakes } from './inventory.js';

export const erpTools = [
  listProducts, getProduct, getProductByBarcode, createProduct, updateProduct, deleteProduct,
  listSuppliers, createSupplier, updateSupplier, deleteSupplier,
  listPurchaseOrders, getPurchaseOrder, createPurchaseOrder, approvePurchaseOrder,
  inventoryReport, listPriceLists, createPriceList, listStockTransfers, listStockTakes,
];
