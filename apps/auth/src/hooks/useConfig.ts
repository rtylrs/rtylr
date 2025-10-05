import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../lib/api';
import { useConfigStore } from '../stores/configStore';

interface ConfigData {
  maintanance: boolean;
  aboutContent: any;
  count: number;
  contactEmail: string;
  contactNumber: string;
  faqs: any[];
  news: any[];
}
export const useConfigData = () => {
  const {
    setAboutContent,
    setMaintanance,
    setBusinessCount,
    setContactInfo,
    setFaqs,
    setNews,
    setLoading,
  } = useConfigStore();

  return useQuery({
    queryKey: ['config'],
    queryFn: async (): Promise<ConfigData> => {
      const response = await api.get('/config');
      return response.data.config;
    },
    onSuccess: (data) => {
      setAboutContent(data.aboutContent);
      setMaintanance(data.maintanance);
      setBusinessCount(data.count);
      setContactInfo({
        contactEmail: data.contactEmail,
        contactNumber: data.contactNumber,
      });
      setFaqs(data.faqs);
      setNews(data.news);
      setLoading(false);
    },
    onError: () => {
      setLoading(false);
      console.error('Error fetching config data');
    },
  });
};

export const useUpdateAboutContent = () => {
  const queryClient = useQueryClient();
  const { setAboutContent } = useConfigStore();

  return useMutation({
    mutationFn: async (data: { aboutContent: string }) => {
      const response = await api.put('/config/update', data);
      return response.data;
    },
    onSuccess: (data) => {
      setAboutContent(data.aboutContent);
      queryClient.invalidateQueries({ queryKey: ['config'] });
    },
    onError: (error) => {
      console.error('Error updating about content:', error);
    },
  });
};

export const useFaqs = () => {
  const { setFaqs } = useConfigStore();

  return useQuery({
    queryKey: ['faqs'],
    queryFn: async () => {
      const response = await api.get('/config/faqs');
      return response.data;
    },
    onSuccess: (data) => {
      setFaqs(data);
    },
    onError: (error) => {
      console.error('Error fetching FAQs:', error);
    },
  });
};
