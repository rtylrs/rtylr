import React from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import api from '../../lib/api';
import { useUserStore } from '../../stores/userStore';

// User info query
export const useUserInfo = () => {
  const { isAuthenticated, setUser, setRole, setUserId } = useUserStore();

  const query = useQuery({
    queryKey: ['userInfo'],
    queryFn: async () => {
      const response = await api.get('/user/info');
      return response.data;
    },
    enabled: isAuthenticated && typeof window !== 'undefined', // Only run on client side when authenticated
  });

  // Handle the data with useEffect
  React.useEffect(() => {
    if (query.data) {
      setUser(query.data);
      setRole(query.data.role);
      setUserId(query.data._id);
    }
  }, [query.data, setUser, setRole, setUserId]);

  return query;
};

// Login mutation
export const useLogin = () => {
  const { setToken, setAuthenticated, setUser, setRole, setUserId } = useUserStore();

  return useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      const response = await api.post('/auth/login', credentials);
      return response.data;
    },
    onSuccess: (data) => {
      if (data.token) {
        setToken(data.token);
        setAuthenticated(true);
        if (data.user) {
          setUser(data.user);
          setRole(data.user.role);
          setUserId(data.user._id);
        }
      }
    },
  });
};
