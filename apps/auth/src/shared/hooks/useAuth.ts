import { useMutation } from '@tanstack/react-query';
import { unwrap, type ApiError } from '@voxire/api';
import { safe } from '../lib/api';
import { useUserStore } from '../stores/userStore';
import { setCookie, getDefaultCookieDomain } from '../lib/cookies';

interface LoginPayload {
  email: string;
  password: string;
}

interface RegisterPayload {
  email: string;
  password: string;
  name: string;
  businessName?: string;
  planSlug?: string;
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
    mutationFn: (payload: LoginPayload): Promise<AuthResponse> =>
      unwrap(safe.post<AuthResponse>('/auth/login', payload)),
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
    onError: (error: ApiError) => {
      setError(error.message || 'Login failed');
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
    mutationFn: (payload: RegisterPayload): Promise<AuthResponse> =>
      unwrap(safe.post<AuthResponse>('/auth/register', payload)),
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
    onError: (error: ApiError) => {
      setError(error.message || 'Registration failed');
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
    mutationFn: (payload: GoogleAuthPayload): Promise<AuthResponse> =>
      unwrap(safe.post<AuthResponse>('/auth/register/google', payload)),
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
    onError: (error: ApiError) => {
      setError(error.message || 'Google login failed');
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};
