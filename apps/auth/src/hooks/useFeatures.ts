import { useQuery } from '@tanstack/react-query';
import { useEffect } from 'react';
import api from '../lib/api';
import { useFeaturesStore } from '../stores/featuresStore';

interface GetFeaturesResponse {
  features: {
    flags: {
      advanced_search: boolean;
      beta_features: boolean;
      dark_mode: boolean;
      google_signup: boolean;
    };
    is_anonymous: boolean;
  };
}

export type FeaturesData = GetFeaturesResponse['features'];

export const useFeatures = () => {
  const { setFlags, setLoading } = useFeaturesStore();

  const query = useQuery<FeaturesData>({
    queryKey: ['features'] as const,
    queryFn: async (): Promise<FeaturesData> => {
      const response = await api.get('/get-features');
      return response.data.features;
    },
  });

  useEffect(() => {
    if (query.data) {
      setFlags(query.data.flags);
      setLoading(false);
    }
  }, [query.data, setFlags, setLoading]);

  useEffect(() => {
    if (query.isError) {
      setLoading(false);
      // eslint-disable-next-line no-console
      console.error('Error fetching features');
    }
  }, [query.isError, setLoading]);

  return query;
};

export const useGoogleSignupVisible = (): boolean => {
  const query = useFeatures();
  const { flags } = useFeaturesStore();
  return query.data?.flags?.google_signup ?? flags.google_signup;
};
