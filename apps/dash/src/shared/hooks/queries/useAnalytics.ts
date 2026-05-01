import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import { safe } from '../../lib/api';
import { analyticsResponseSchema } from '../../schemas/analytics';
import { toast } from 'sonner';

export const useGetAnalytics = (
  businessId: string | null,
  dateRange?: { start: string; end: string }
) => {
  const query = useQuery({
    queryKey: ['analytics', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await unwrap(safe.get(`/business/${businessId}/analytics`, { params }));
      return analyticsResponseSchema.parse(response.data);
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch analytics');
    }
  }, [query.error]);

  return query;
};

export const useGetRevenue = (
  businessId: string | null,
  dateRange?: { start: string; end: string }
) => {
  const query = useQuery({
    queryKey: ['revenue', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await unwrap(
        safe.get(`/business/${businessId}/analytics/revenue`, { params })
      );
      return analyticsResponseSchema.parse(response.data);
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
    }
  }, [query.error]);

  return query;
};

export const useGetOrders = (
  businessId: string | null,
  dateRange?: { start: string; end: string }
) => {
  const query = useQuery({
    queryKey: ['orders', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await unwrap(
        safe.get(`/business/${businessId}/analytics/orders`, { params })
      );
      return analyticsResponseSchema.parse(response.data);
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
    }
  }, [query.error]);

  return query;
};
