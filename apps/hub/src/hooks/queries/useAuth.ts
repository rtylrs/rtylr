import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import api from '../../lib/api';
import { useUserStore } from '../../stores/userStore';
import { toast } from 'sonner';

interface LoginPayload {
  email: string;
  password: string;
}

interface RegisterPayload {
  email: string;
  password: string;
  name: string;
  referred: string;
}

interface GoogleAuthPayload {
  idToken: string;
}

export const useLogin = () => {
  const { setToken, setUser, setRole, setUserId, setAddons, setVisits } = useUserStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: LoginPayload) => {
      const response = await api.post('/auth/login', payload);
      return response.data;
    },
    onSuccess: (data) => {
      localStorage.setItem('token', data.token);
      setToken(data.token);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || error.response?.data?.error || 'Login failed';
      toast.error(message);
      throw error;
    },
  });
};

export const useRegister = () => {
  const { setToken } = useUserStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: RegisterPayload) => {
      const response = await api.post('/auth/create/business', payload);
      return response.data;
    },
    onSuccess: (data) => {
      localStorage.setItem('token', data.token);
      setToken(data.token);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || error.response?.data?.error || 'Registration failed';
      toast.error(message);
      throw error;
    },
  });
};

export const useGoogleAuth = () => {
  const { setToken } = useUserStore();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (payload: GoogleAuthPayload) => {
      const response = await api.post('/auth/create/business/google', payload);
      return response.data;
    },
    onSuccess: (data) => {
      localStorage.setItem('token', data.token);
      setToken(data.token);
      queryClient.invalidateQueries({ queryKey: ['user'] });
    },
    onError: (error: any) => {
      const message = error.response?.data?.message || error.response?.data?.error || 'Google login failed';
      toast.error(message);
      throw error;
    },
  });
};

export const useGetUserInfo = () => {
  const { setUser, setRole, setUserId, setToken, setAddons, setVisits } = useUserStore();

  const query = useQuery({
    queryKey: ['user'],
    queryFn: async () => {
      const response = await api.get('/business/me');
      return response.data;
    },
    enabled: !!localStorage.getItem('token'),
  });

  useEffect(() => {
    if (query.data) {
      const { userId, token, user, visits } = query.data;
      setRole(user.role);
      setUserId(userId);
      setToken(token);
      setUser(user);
      setAddons(user.addons);
      setVisits(visits);
    }
  }, [query.data, setUser, setRole, setUserId, setToken, setAddons, setVisits]);

  useEffect(() => {
    if (query.error) {
      const error: any = query.error;
      const message = error.response?.data?.message || error.response?.data?.error || 'Failed to get user info';
      toast.error(message);
    }
  }, [query.error]);

  return query;
};

