import React from 'react';
import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';
import { useConfigStore } from '../../stores/configStore';

// Config data query
export const useConfigData = () => {
  const {
    setWebsiteData,
    setAboutContent,
    setMaintanance,
    setBusinessCount,
    setContactInfo,
    setLoading,
    setError
  } = useConfigStore();

  const query = useQuery({
    queryKey: ['configData'],
    queryFn: async () => {
      const response = await api.get('/config');
      return response.data;
    },
    enabled: typeof window !== 'undefined', // Only run on client side
  });

  // Handle the data with useEffect
  React.useEffect(() => {
    if (query.isLoading) {
      setLoading(true);
    } else {
      setLoading(false);
    }

    if (query.error) {
      setError((query.error as any).message || 'Failed to fetch config data');
    } else {
      setError(null);
    }

    if (query.data) {
      setWebsiteData(query.data.websiteData);
      setAboutContent(query.data.aboutContent);
      setMaintanance(query.data.maintanance || false);
      setBusinessCount(query.data.businessCount || 0);
      setContactInfo(query.data.contactInfo || { contactEmail: null, contactNumber: null });
    }
  }, [query.data, query.isLoading, query.error, setWebsiteData, setAboutContent, setMaintanance, setBusinessCount, setContactInfo, setLoading, setError]);

  return query;
};

// FAQs query
export const useFaqs = () => {
  const { setFaqs } = useConfigStore();

  const query = useQuery({
    queryKey: ['faqs'],
    queryFn: async () => {
      const response = await api.get('/faqs');
      return response.data;
    },
    enabled: typeof window !== 'undefined', // Only run on client side
  });

  React.useEffect(() => {
    if (query.data) {
      setFaqs(query.data || []);
    }
  }, [query.data, setFaqs]);

  return query;
};

// News query
export const useNews = () => {
  const { setNews } = useConfigStore();

  const query = useQuery({
    queryKey: ['news'],
    queryFn: async () => {
      const response = await api.get('/news');
      return response.data;
    },
    enabled: typeof window !== 'undefined', // Only run on client side
  });

  React.useEffect(() => {
    if (query.data) {
      setNews(query.data || []);
    }
  }, [query.data, setNews]);

  return query;
};