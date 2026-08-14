import { createApiClient, safeFetch } from '@voxire/api';
import type { SafeApi } from '@voxire/api';
import type { AxiosRequestConfig } from 'axios';

// Recruit talks to the same backend as every other app, but its public
// pages are anonymous - the axios client still attaches the cookie if
// one happens to be present (so an admin previewing their own page
// doesn't look logged out), and the backend treats /recruit/public/*
// as anonymous regardless.
const api = createApiClient('recruit');
const rawSafe = safeFetch(api);

const camelKey = (key: string): string => key.replace(/_([a-z])/g, (_, letter: string) => letter.toUpperCase());
const snakeKey = (key: string): string => key.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);

export const camelizeRecruitPayload = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(camelizeRecruitPayload);
  if (!value || typeof value !== 'object') return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [camelKey(key), camelizeRecruitPayload(item)])
  );
};

export const snakeizeRecruitPayload = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(snakeizeRecruitPayload);
  if (!value || typeof value !== 'object') return value;
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [snakeKey(key), snakeizeRecruitPayload(item)])
  );
};

const snakeConfig = (config?: AxiosRequestConfig): AxiosRequestConfig | undefined => {
  if (!config) return config;
  return {
    ...config,
    data: config.data === undefined ? undefined : snakeizeRecruitPayload(config.data),
  };
};

export const safe: SafeApi = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    rawSafe.get<unknown>(url, config).map((value) => camelizeRecruitPayload(value) as T),
  post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    rawSafe
      .post<unknown>(url, snakeizeRecruitPayload(data), config)
      .map((value) => camelizeRecruitPayload(value) as T),
  put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    rawSafe
      .put<unknown>(url, snakeizeRecruitPayload(data), config)
      .map((value) => camelizeRecruitPayload(value) as T),
  patch: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
    rawSafe
      .patch<unknown>(url, snakeizeRecruitPayload(data), config)
      .map((value) => camelizeRecruitPayload(value) as T),
  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    rawSafe
      .delete<unknown>(url, snakeConfig(config))
      .map((value) => camelizeRecruitPayload(value) as T),
};
export default api;
