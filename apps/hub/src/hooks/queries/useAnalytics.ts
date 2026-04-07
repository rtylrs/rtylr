import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import api from '../../lib/api';
import { toast } from 'sonner';

export const useGetAnalytics = (businessId: string | null, dateRange?: { start: string; end: string }) => {
  const query = useQuery({
    queryKey: ['analytics', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await api.get(`/business/${businessId}/analytics`, { params });
      return response.data;
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch analytics');
      console.error('Error fetching analytics:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useGetRevenue = (businessId: string | null, dateRange?: { start: string; end: string }) => {
  const query = useQuery({
    queryKey: ['revenue', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await api.get(`/business/${businessId}/analytics/revenue`, { params });
      return response.data;
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
      console.error('Error fetching revenue:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useGetOrders = (businessId: string | null, dateRange?: { start: string; end: string }) => {
  const query = useQuery({
    queryKey: ['orders', businessId, dateRange],
    queryFn: async () => {
      const params = dateRange ? { start: dateRange.start, end: dateRange.end } : {};
      const response = await api.get(`/business/${businessId}/analytics/orders`, { params });
      return response.data;
    },
    enabled: !!businessId,
  });

  useEffect(() => {
    if (query.error) {
      console.error('Error fetching orders:', query.error);
    }
  }, [query.error]);

  return query;
};

