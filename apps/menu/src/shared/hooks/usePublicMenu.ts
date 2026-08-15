import { useQuery } from '@tanstack/react-query';
import { cdnImageUrl } from '@voxire/config';
import { z } from 'zod';
import api from '../lib/api';

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

interface BusinessApiResponse {
  id: number;
  name: string;
  logo_url: string;
  currency: string;
  timezone: string;
}

interface MenuApiResponse {
  items: {
    id: number;
    owner_id: number;
    category_id: number;
    name: string;
    sku: string;
    sale_price: number;
    image?: string;
    image_url?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
  }[];
}

const publicBusinessSchema = z.object({
  id: z.number().catch(0),
  name: z.string().catch(''),
  logo_url: z.string().catch(''),
  currency: z.string().catch('USD'),
  timezone: z.string().catch('UTC'),
});

const publicMenuItemSchema = z.object({
  id: z.number().catch(0),
  owner_id: z.number().catch(0),
  category_id: z.number().catch(0),
  name: z.string().catch(''),
  sku: z.string().catch(''),
  sale_price: z.number().catch(0),
  image: z.string().optional(),
  image_url: z.string().optional(),
  is_active: z.boolean().catch(true),
  created_at: z.string().catch(''),
  updated_at: z.string().catch(''),
});

async function fetchPublicBusiness(businessId: number): Promise<PublicBusiness> {
  const res = await api.get<{ business: BusinessApiResponse }>(
    '/hub/business/v1/businesses/public',
    { params: { business_id: businessId } }
  );
  const data = publicBusinessSchema.parse(res.data.business ?? {});
  return {
    id: data.id,
    name: data.name,
    logoUrl: cdnImageUrl(data.logo_url) ?? '',
    currency: data.currency,
    timezone: data.timezone,
  };
}

async function fetchPublicMenu(
  businessId: number,
  locationId?: number | null
): Promise<PublicMenuItem[]> {
  const params: Record<string, number> = { business_id: businessId };
  if (locationId && locationId > 0) params.location_id = locationId;
  const res = await api.get<MenuApiResponse>('/pos/menu-item/v1/menu-items/public', { params });
  const items = (res.data.items ?? []).map((item) => publicMenuItemSchema.parse(item));
  return items.map((i) => ({
    id: i.id,
    ownerId: i.owner_id,
    categoryId: i.category_id,
    name: i.name,
    sku: i.sku,
    salePrice: i.sale_price,
    imageUrl: cdnImageUrl(i.image_url || i.image),
    isActive: i.is_active,
    createdAt: i.created_at,
    updatedAt: i.updated_at,
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

interface PublicBusinessBySlugApi {
  business: {
    id: number;
    name: string;
    slug: string;
    logo_url: string;
    currency: string;
    timezone: string;
  };
  locations: {
    id: number;
    name: string;
    slug: string;
    address: string;
    timezone: string;
    is_active: boolean;
  }[];
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
      const res = await api.get<PublicBusinessBySlugApi>('/hub/business/v1/businesses/by-slug', {
        params: { slug },
      });
      return {
        business: {
          id: res.data.business.id,
          name: res.data.business.name,
          slug: res.data.business.slug,
          logoUrl: cdnImageUrl(res.data.business.logo_url) ?? '',
          currency: res.data.business.currency,
          timezone: res.data.business.timezone,
        },
        locations: (res.data.locations ?? []).map((l) => ({
          id: l.id,
          name: l.name,
          slug: l.slug,
          address: l.address,
          timezone: l.timezone,
          isActive: l.is_active,
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
      // Two-trip parallel. If either fails the whole bundle fails -
      // a menu without a business header renders poorly and vice
      // versa, so we'd rather show the error state once than half-
      // render.
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
