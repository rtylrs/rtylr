import type { AddonData } from '../../shared/hooks/queries/useAddons';
import { INTEGRATION_CATALOG, type Integration } from '../../shared/schemas/integrations';

export const EXTENSION_CATEGORIES = [
  'all',
  'commerce',
  'marketing',
  'business',
  'payments',
  'messaging',
  'accounting',
  'analytics',
  'delivery',
  'data',
] as const;
export type ExtensionCategory = (typeof EXTENSION_CATEGORIES)[number];
export type ExtensionEntry =
  | {
      kind: 'feature';
      key: string;
      name: string;
      description: string;
      category: ExtensionCategory;
      added: boolean;
      addon: AddonData;
    }
  | {
      kind: 'connection';
      key: string;
      name: string;
      description: string;
      category: ExtensionCategory;
      added: boolean;
      integration: Integration;
    };

const BUILT_IN_ITEMS = new Set(['insights', 'docs', 'documentation', 'ai_chat']);

const FEATURE_CATEGORIES: Record<string, ExtensionCategory> = {
  menu: 'commerce',
  loyalty: 'commerce',
  store: 'commerce',
  book: 'commerce',
  reservations: 'commerce',
  online_ordering: 'commerce',
  kds: 'commerce',
  email_marketing: 'marketing',
  sms: 'messaging',
  sales_pipeline: 'marketing',
  feedback: 'marketing',
  about: 'marketing',
  contact: 'marketing',
  landing: 'marketing',
  analytics: 'analytics',
  insights: 'analytics',
  data_export: 'data',
  api_access: 'data',
};

export function buildExtensionCatalog(
  addons: AddonData[],
  connections: Integration[]
): ExtensionEntry[] {
  const catalog: ExtensionEntry[] = addons
    .filter((addon) => !BUILT_IN_ITEMS.has(addon.id))
    .map((addon) => ({
      kind: 'feature',
      key: `feature:${addon.id}`,
      name: addon.name,
      description: addon.description,
      category: FEATURE_CATEGORIES[addon.id] ?? 'business',
      added: addon.isEnabled,
      addon,
    }));
  const providers = new Set(connections.map((row) => row.provider.toLowerCase()));
  const integrations: Integration[] = [
    ...connections,
    ...INTEGRATION_CATALOG.filter((row) => !providers.has(row.provider)).map((row) => ({
      ...row,
      id: `catalog:${row.provider}`,
      logo: row.provider,
      status: 'available',
      config: undefined,
      connectedAt: undefined,
      createdAt: undefined,
    })),
  ];
  return catalog.concat(
    integrations.map((integration) => ({
      kind: 'connection' as const,
      key: `connection:${integration.id}`,
      name: integration.name,
      description: integration.description,
      category: integration.category,
      added: integration.status === 'connected',
      integration,
    }))
  );
}

export function getFeatureSettings(id: string): '/settings/domain' | '/settings/email' | undefined {
  if (id === 'custom_domain') return '/settings/domain';
  if (id === 'email_marketing') return '/settings/email';
  return undefined;
}
