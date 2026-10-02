import { useQuery } from '@tanstack/react-query';
import { cdnImageUrl } from '@voxire/config';
import { businessClient, menuItemClient } from '../lib/clients';

export interface PublicBusiness {
  id: number;
  name: string;
  logoUrl: string;
  currency: string;
  timezone: string;
}

export interface PublicMenuItem {
  id: number;
  ownerId: number;
  categoryId: number;
  name: string;
  sku: string;
  salePrice: number;
  imageUrl?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

const toNumber = (value: string | undefined, fallback = 0): number => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

async function fetchPublicBusiness(businessId: number): Promise<PublicBusiness> {
  const response = await businessClient.getPublicBusiness({ businessId: String(businessId) });
  const business = response.business;
  if (!business) throw new Error('Public business was not returned');

  return {
    id: toNumber(business.id, businessId),
    name: business.name ?? '',
    logoUrl: business.logoUrl ? (cdnImageUrl(business.logoUrl) ?? '') : '',
    currency: business.currency ?? 'USD',
    timezone: business.timezone ?? 'UTC',
  };
}

async function fetchPublicMenu(
  businessId: number,
  locationId?: number | null
): Promise<PublicMenuItem[]> {
  const response = await menuItemClient.listPublicMenuItems({
    businessId: String(businessId),
    ...(locationId && locationId > 0 ? { locationId: String(locationId) } : {}),
  });

  return (response.items ?? []).map((item) => ({
    id: toNumber(item.id),
    ownerId: toNumber(item.ownerId),
    categoryId: toNumber(item.categoryId),
    name: item.name ?? '',
    sku: item.sku ?? '',
    salePrice: item.salePrice ?? 0,
    imageUrl: item.imageUrl ? (cdnImageUrl(item.imageUrl) ?? undefined) : undefined,
    isActive: item.isActive ?? true,
    createdAt: item.createdAt ?? '',
    updatedAt: item.updatedAt ?? '',
  }));
}

export interface PublicMenuBundle {
  business: PublicBusiness;
  items: PublicMenuItem[];
}

export interface PublicLocationSummary {
  id: number;
  name: string;
  slug: string;
  address: string;
  timezone: string;
  isActive: boolean;
}

export interface PublicBusinessBySlugBundle {
  business: PublicBusiness & { slug: string };
  locations: PublicLocationSummary[];
}

export function usePublicBusinessBySlug(slug: string | null): {
  data: PublicBusinessBySlugBundle | undefined;
  isLoading: boolean;
  isError: boolean;
} {
  const query = useQuery({
    queryKey: ['public-business-by-slug', slug],
    queryFn: async (): Promise<PublicBusinessBySlugBundle> => {
      if (!slug) throw new Error('slug required');

      const response = await businessClient.lookupBusinessBySlug({ slug });
      const business = response.business;
      if (!business) throw new Error('Public business was not returned');

      return {
        business: {
          id: toNumber(business.id),
          name: business.name ?? '',
          slug: business.slug ?? slug,
          logoUrl: business.logoUrl ? (cdnImageUrl(business.logoUrl) ?? '') : '',
          currency: business.currency ?? 'USD',
          timezone: business.timezone ?? 'UTC',
        },
        locations: (response.locations ?? []).map((location) => ({
          id: toNumber(location.id),
          name: location.name ?? '',
          slug: location.slug ?? '',
          address: location.address ?? '',
          timezone: location.timezone ?? 'UTC',
          isActive: location.isActive ?? true,
        })),
      };
    },
    enabled: !!slug,
    staleTime: 60_000,
  });

  return {
    data: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
  };
}

export function usePublicMenu(
  businessId: number | null,
  locationId?: number | null
): {
  data: PublicMenuBundle | undefined;
  isLoading: boolean;
  isError: boolean;
} {
  const query = useQuery({
    queryKey: ['public-menu', businessId, locationId ?? null],
    queryFn: async (): Promise<PublicMenuBundle> => {
      if (businessId == null) {
        throw new Error('business_id required');
      }

      // Two-trip parallel. If either fails the whole bundle fails - a menu
      // without a business header renders poorly and vice versa.
      const [business, items] = await Promise.all([
        fetchPublicBusiness(businessId),
        fetchPublicMenu(businessId, locationId),
      ]);
      return { business, items };
    },
    enabled: businessId != null && Number.isFinite(businessId),
    staleTime: 60_000,
  });

  return {
    data: query.data,
    isLoading: query.isLoading,
    isError: query.isError,
  };
}
