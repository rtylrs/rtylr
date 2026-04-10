import { useMutation } from '@tanstack/react-query';
import api from '../lib/api';
import { useUserStore } from '../stores/userStore';
import {
  setCookie,
  getDefaultCookieDomain,
} from '../lib/cookies';

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

interface AuthResponse {
  token: string;
  role: string;
}

export const useLogin = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation({
    mutationFn: async (payload: LoginPayload): Promise<AuthResponse> => {
      const response = await api.post('/auth/login', payload);
      return response.data;
    },
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data) => {
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });
      login(data.role);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message ||
                          error.response?.data?.error ||
                          'Login failed';
      setError(errorMessage);
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const useRegisterRestaurant = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation({
    mutationFn: async (payload: RegisterPayload): Promise<AuthResponse> => {
      const response = await api.post('/auth/register', payload);
      return response.data;
    },
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data) => {
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });
      login(data.role);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message ||
                          error.response?.data?.error ||
                          'Registration failed';
      setError(errorMessage);
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const useGoogleAuth = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation({
    mutationFn: async (payload: GoogleAuthPayload): Promise<AuthResponse> => {
      const response = await api.post('/auth/register/google', payload);
      return response.data;
    },
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data) => {
      setCookie('token', data.token, {
        days: 30,
        sameSite: 'lax',
        domain: getDefaultCookieDomain(),
      });
      login(data.role);
    },
    onError: (error: any) => {
      const errorMessage = error.response?.data?.message ||
                          error.response?.data?.error ||
                          'Google login failed';
      setError(errorMessage);
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};