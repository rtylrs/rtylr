import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { safe } from '../../lib/api';
import { useConfigStore } from '../../stores/configStore';
import { toast } from 'sonner';

export const useGetConfig = () => {
  const {
    setAboutContent,
    setMaintanance,
    setBusinessCount,
    setContactInfo,
    setFaqs,
    setNews,
  } = useConfigStore();

  const query = useQuery({
    queryKey: ['config'],
    queryFn: async () => {
      const response = await unwrap(safe.get('/config'));
      return response.data.config;
    },
  });

  useEffect(() => {
    if (query.data) {
      const {
        maintanance,
        aboutContent,
        count,
        contactEmail,
        contactNumber,
        faqs,
        news,
      } = query.data;

      setAboutContent(aboutContent);
      setMaintanance(maintanance);
      setBusinessCount(count);
      setContactInfo({
        contactEmail,
        contactNumber,
      });
      setFaqs(faqs);
      setNews(news);
    }
  }, [query.data, setAboutContent, setMaintanance, setBusinessCount, setContactInfo, setFaqs, setNews]);

  useEffect(() => {
    if (query.error) {
      console.error('Error fetching config data:', query.error);
    }
  }, [query.error]);

  return query;
};

export const useUpdateConfig = () => {
  const queryClient = useQueryClient();
  const { setAboutContent } = useConfigStore();

  return useMutation({
    mutationFn: async (data: { aboutContent: string }) => {
      const response = await unwrap(safe.put('/config/update', data));
      return response.data;
    },
    onSuccess: (data) => {
      setAboutContent(data.aboutContent);
      queryClient.invalidateQueries({ queryKey: ['config'] });
      toast.success('Config updated successfully');
    },
    onError: (error) => {
      toast.error('Failed to update config');
      console.error('Error updating about content:', error);
    },
  });
};

export const useGetFaqs = () => {
  const { setFaqs } = useConfigStore();

  const query = useQuery({
    queryKey: ['faqs'],
    queryFn: async () => {
      const response = await unwrap(safe.get('/config/faqs'));
      return response.data;
    },
  });

  useEffect(() => {
    if (query.data) {
      setFaqs(query.data);
    }
  }, [query.data, setFaqs]);

  useEffect(() => {
    if (query.error) {
      toast.error('Failed to fetch FAQs');
      console.error('Error fetching FAQs:', query.error);
    }
  }, [query.error]);

  return query;
};

