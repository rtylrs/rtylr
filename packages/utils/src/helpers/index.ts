import { debounce, throttle, isEmpty, isEqual } from 'lodash';

export const helpers = {
  // Debounce function calls
  debounce: <T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): ((...args: Parameters<T>) => void) => {
    return debounce(func, wait);
  },

  // Throttle function calls
  throttle: <T extends (...args: any[]) => any>(
    func: T,
    wait: number
  ): ((...args: Parameters<T>) => void) => {
    return throttle(func, wait);
  },

  // Check if value is empty
  isEmpty: (value: any): boolean => {
    return isEmpty(value);
  },

  // Deep equality check
  isEqual: (value1: any, value2: any): boolean => {
    return isEqual(value1, value2);
  },

  // Generate random ID
  generateId: (length: number = 8): string => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  },

  // Deep clone object
  deepClone: <T>(obj: T): T => {
    if (obj === null || typeof obj !== 'object') return obj;
    if (obj instanceof Date) return new Date(obj.getTime()) as T;
    if (obj instanceof RegExp) return new RegExp(obj.source, obj.flags) as T;
    if (obj instanceof Set)
      return new Set(Array.from(obj).map((item) => helpers.deepClone(item))) as T;
    if (obj instanceof Map)
      return new Map(
        Array.from(obj.entries()).map(([k, v]) => [helpers.deepClone(k), helpers.deepClone(v)])
      ) as T;
    if (obj instanceof Array) return obj.map((item) => helpers.deepClone(item)) as T;
    if (typeof obj === 'object') {
      const clonedObj = {} as T;
      for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
          clonedObj[key] = helpers.deepClone(obj[key]);
        }
      }
      return clonedObj;
    }
    return obj;
  },

  // Retry function with exponential backoff
  retry: async <T>(
    fn: () => Promise<T>,
    maxAttempts: number = 3,
    delay: number = 1000
  ): Promise<T> => {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await fn();
      } catch (error) {
        lastError = error as Error;
        if (attempt === maxAttempts) break;

        await new Promise((resolve) => setTimeout(resolve, delay * Math.pow(2, attempt - 1)));
      }
    }

    throw lastError!;
  },

  // Sleep utility
  sleep: (ms: number): Promise<void> => {
    return new Promise((resolve) => setTimeout(resolve, ms));
  },

  // Get nested object property safely (guarded against prototype pollution)
  get: (obj: any, path: string, defaultValue?: any): any => {
    if (obj == null) return defaultValue;
    const keys = path.split('.');
    let result = obj;

    for (const key of keys) {
      if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
        return defaultValue;
      }
      if (result == null || typeof result !== 'object') {
        return defaultValue;
      }
      result = result[key];
    }

    return result !== undefined ? result : defaultValue;
  },

  // Set nested object property safely (guarded against prototype pollution)
  set: (obj: any, path: string, value: any): void => {
    if (obj == null || typeof obj !== 'object') return;
    const keys = path.split('.');
    let current = obj;

    for (let i = 0; i < keys.length - 1; i++) {
      const key = keys[i];
      if (key === '__proto__' || key === 'constructor' || key === 'prototype') {
        return;
      }
      if (!(key in current) || current[key] == null || typeof current[key] !== 'object') {
        current[key] = {};
      }
      current = current[key];
    }

    const lastKey = keys[keys.length - 1];
    if (lastKey === '__proto__' || lastKey === 'constructor' || lastKey === 'prototype') {
      return;
    }
    current[lastKey] = value;
  },
};

export default helpers;
