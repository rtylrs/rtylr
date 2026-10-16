import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useCustomDomain,
  useRegisterCustomDomain,
  useRemoveCustomDomain,
  useUpdateBusiness,
  useVerifyCustomDomain,
} from './useBusiness';
import { useConfigStore } from '../../stores/configStore';

const mockConfigClient = vi.hoisted(() => ({
  updateConfig: vi.fn(),
  getCustomDomain: vi.fn(),
  registerCustomDomain: vi.fn(),
  verifyCustomDomain: vi.fn(),
  removeCustomDomain: vi.fn(),
}));
const mockToastSuccess = vi.fn();
const mockToastError = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
  getOneKitErrorMessage: vi.fn((_err, fallback) => fallback),
}));

vi.mock('../../lib/clients', () => ({ configClient: mockConfigClient }));

vi.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useUpdateBusiness', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useConfigStore.setState({
      websiteData: null,
      aboutContent: null,
      maintanance: false,
      businessCount: 0,
      contactInfo: { contactEmail: null, contactNumber: null },
      faqs: [],
      news: [],
    });
  });

  it('posts config updates using the backend snake_case contract', async () => {
    mockConfigClient.updateConfig.mockResolvedValue({
      config: {
        site_name: 'Suplab Store',
        contact_email: 'suplabstore@gmail.com',
        contact_number: '+961000000',
        sender_email: 'orders@suplabstore.com',
        maps_url: 'https://maps.example.com',
        website: 'https://suplabstore.com/',
      },
    });

    const { result } = renderHook(() => useUpdateBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate({
        data: {
          storeName: 'Suplab Store',
          contactEmail: 'suplabstore@gmail.com',
          telephone: '+961000000',
          senderEmail: 'orders@suplabstore.com',
          mapsUrl: 'https://maps.example.com',
          website: 'https://suplabstore.com/',
        },
      });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockConfigClient.updateConfig).toHaveBeenCalledWith({
      siteName: 'Suplab Store',
      contactEmail: 'suplabstore@gmail.com',
      contactNumber: '+961000000',
      senderEmail: 'orders@suplabstore.com',
      mapsUrl: 'https://maps.example.com',
      website: 'https://suplabstore.com/',
    });
    expect(useConfigStore.getState().websiteData).toMatchObject({
      siteName: 'Suplab Store',
      contactEmail: 'suplabstore@gmail.com',
      contactNumber: '+961000000',
      senderEmail: 'orders@suplabstore.com',
      mapsUrl: 'https://maps.example.com',
      website: 'https://suplabstore.com/',
    });
    expect(useConfigStore.getState().contactInfo).toEqual({
      contactEmail: 'suplabstore@gmail.com',
      contactNumber: '+961000000',
    });
  });

  it('forwards an explicitly empty store name', async () => {
    mockConfigClient.updateConfig.mockResolvedValue({ config: { site_name: '' } });

    const { result } = renderHook(() => useUpdateBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate({ data: { storeName: '', companyName: 'Old name' } });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockConfigClient.updateConfig).toHaveBeenCalledWith({ siteName: '' });
  });

  it('uses company name only when store name was not supplied', async () => {
    mockConfigClient.updateConfig.mockResolvedValue({ config: { site_name: 'Company name' } });

    const { result } = renderHook(() => useUpdateBusiness(), { wrapper: createWrapper() });

    act(() => {
      result.current.mutate({ data: { companyName: 'Company name' } });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockConfigClient.updateConfig).toHaveBeenCalledWith({ siteName: 'Company name' });
  });

  it('invalidates cached HR setup after changing the shared company name', async () => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    client.setQueryData(['hr', 'hr-setup-status'], { config: { companyName: 'Old name' } });
    mockConfigClient.updateConfig.mockResolvedValue({ config: { site_name: 'New name' } });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useUpdateBusiness(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ data: { companyName: 'New name' } });
    });
    expect(client.getQueryState(['hr', 'hr-setup-status'])?.isInvalidated).toBe(true);
    expect(useConfigStore.getState().websiteData?.siteName).toBe('New name');
  });
});

describe('useCustomDomain', () => {
  beforeEach(() => vi.clearAllMocks());

  it('reads the registered domain record', async () => {
    mockConfigClient.getCustomDomain.mockResolvedValue({
      domain: { domain: 'menu.acme.com', status: 'pending' },
    });

    const { result } = renderHook(() => useCustomDomain(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toMatchObject({ domain: 'menu.acme.com', status: 'pending' });
  });
});

describe('useRegisterCustomDomain', () => {
  beforeEach(() => vi.clearAllMocks());

  it('registers through the custom-domain RPC, not the config update', async () => {
    mockConfigClient.registerCustomDomain.mockResolvedValue({
      domain: { domain: 'menu.acme.com' },
    });

    const { result } = renderHook(() => useRegisterCustomDomain(), { wrapper: createWrapper() });
    act(() => {
      result.current.mutate('menu.acme.com');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockConfigClient.registerCustomDomain).toHaveBeenCalledWith({ domain: 'menu.acme.com' });
    expect(mockConfigClient.updateConfig).not.toHaveBeenCalled();
    expect(mockToastSuccess).toHaveBeenCalled();
  });
});

describe('useVerifyCustomDomain', () => {
  beforeEach(() => vi.clearAllMocks());

  it('sends the challenge token and refreshes the domain record', async () => {
    mockConfigClient.verifyCustomDomain.mockResolvedValue({ domain: { status: 'verified' } });

    const { result } = renderHook(() => useVerifyCustomDomain(), { wrapper: createWrapper() });
    act(() => {
      result.current.mutate('challenge-token');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockConfigClient.verifyCustomDomain).toHaveBeenCalledWith({
      verificationToken: 'challenge-token',
    });
    expect(mockToastSuccess).toHaveBeenCalled();
  });

  it('surfaces a failed verification instead of reporting success', async () => {
    mockConfigClient.verifyCustomDomain.mockRejectedValue(new Error('dns mismatch'));

    const { result } = renderHook(() => useVerifyCustomDomain(), { wrapper: createWrapper() });
    act(() => {
      result.current.mutate('challenge-token');
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(mockToastError).toHaveBeenCalled();
    expect(mockToastSuccess).not.toHaveBeenCalled();
  });
});

describe('useRemoveCustomDomain', () => {
  beforeEach(() => vi.clearAllMocks());

  it('removes the claim through the custom-domain RPC', async () => {
    mockConfigClient.removeCustomDomain.mockResolvedValue({ message: 'ok' });

    const { result } = renderHook(() => useRemoveCustomDomain(), { wrapper: createWrapper() });
    act(() => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockConfigClient.removeCustomDomain).toHaveBeenCalledWith({});
  });
});
