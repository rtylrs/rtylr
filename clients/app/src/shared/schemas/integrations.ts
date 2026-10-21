import { z } from 'zod';

export type IntegrationCategory =
  'payments' | 'messaging' | 'accounting' | 'analytics' | 'delivery' | 'data';

type ProviderDetails = {
  name: string;
  category: IntegrationCategory;
  description: string;
};

const PROVIDER_DETAILS: Record<string, ProviderDetails> = {
  stripe: {
    name: 'Stripe',
    category: 'payments',
    description: 'Accept online payments with Stripe.',
  },
};

const optionalText = z
  .string()
  .nullable()
  .optional()
  .transform((value) => value ?? undefined);

function providerDetails(provider: string): ProviderDetails {
  const normalized = provider.trim().toLowerCase();
  const known = PROVIDER_DETAILS[normalized];
  if (known) return known;

  const name = normalized
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
  return {
    name: name || 'Unknown integration',
    category: 'data',
    description: '',
  };
}

function normalizeStatus(value: string | null | undefined): string {
  switch (value) {
    case 'connected':
    case 'disconnected':
    case 'error':
    case 'available':
    case 'coming_soon':
      return value;
    default:
      return 'disconnected';
  }
}

export const integrationSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String).catch(''),
    provider: z.string().trim().min(1).catch('unknown'),
    status: z.string().nullable().optional().catch('disconnected'),
    config: z.unknown().nullish(),
    connectedAt: optionalText,
    createdAt: optionalText,
  })
  .transform((raw) => {
    const provider = raw.provider.trim() || 'unknown';
    const details = providerDetails(provider);
    return {
      id: raw.id,
      provider,
      name: details.name,
      category: details.category,
      logo: provider,
      status: normalizeStatus(raw.status),
      description: details.description,
      config: raw.config ?? undefined,
      connectedAt: raw.connectedAt,
      createdAt: raw.createdAt,
    };
  });

export type Integration = z.infer<typeof integrationSchema>;

export type IntegrationCatalogEntry = ProviderDetails & {
  provider: string;
};

/**
 * The current API lists connected rows only. Keep the small set of products
 * documented by the Hub available in the UI until the service exposes a
 * provider catalogue of its own.
 */
export const INTEGRATION_CATALOG: readonly IntegrationCatalogEntry[] = [
  {
    provider: 'stripe',
    name: 'Stripe',
    category: 'payments',
    description: 'Accept online payments with Stripe.',
  },
];
