import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const productsBase = '/api/erp/product/v1/products';

export const listProducts: Tool = {
  name: 'erp_list_products',
  description: 'List ERP products / inventory items.',
  inputSchema: {
    type: 'object',
    properties: {
      search: { type: 'string', description: 'Search by name or SKU' },
      category_id: { type: 'string' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`${productsBase}${qs(input, ['search', 'category_id', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getProduct: Tool = {
  name: 'erp_get_product',
  description: 'Get a single ERP product by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      product_id: { type: 'string', description: 'Product ID' },
    },
    required: ['product_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`${productsBase}/${encodeURIComponent(String(input.product_id))}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getProductByBarcode: Tool = {
  name: 'erp_get_product_by_barcode',
  description: 'Look up an ERP product by barcode.',
  inputSchema: {
    type: 'object',
    properties: {
      barcode: { type: 'string', description: 'Barcode value' },
    },
    required: ['barcode'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`${productsBase}/barcode/${encodeURIComponent(String(input.barcode))}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createProduct: Tool = {
  name: 'erp_create_product',
  description: 'Create a new ERP product.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Product name' },
      sku: { type: 'string', description: 'Stock keeping unit' },
      barcode: { type: 'string' },
      price: { type: 'string', description: 'Price in smallest currency unit' },
      cost: { type: 'string', description: 'Cost in smallest currency unit' },
      category_id: { type: 'string' },
      description: { type: 'string' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(productsBase, input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateProduct: Tool = {
  name: 'erp_update_product',
  description: 'Update an ERP product.',
  inputSchema: {
    type: 'object',
    properties: {
      product_id: { type: 'string', description: 'Product ID' },
      name: { type: 'string' },
      sku: { type: 'string' },
      price: { type: 'string' },
      cost: { type: 'string' },
      description: { type: 'string' },
    },
    required: ['product_id'],
  },
  handler: async (input, api) => {
    try {
      const { product_id, ...body } = input;
      const res = await api.put(`${productsBase}/${encodeURIComponent(String(product_id))}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteProduct: Tool = {
  name: 'erp_delete_product',
  description: 'Delete an ERP product.',
  inputSchema: {
    type: 'object',
    properties: {
      product_id: { type: 'string', description: 'Product ID' },
    },
    required: ['product_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`${productsBase}/${encodeURIComponent(String(input.product_id))}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
