import { useMutation } from '@tanstack/react-query';
import { unwrap, type ApiError } from '@voxire/api';
import { validateContractPayload } from '@voxire/contracts/validation/metadata';
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
  code: string;
}

interface AuthResponse {
  token: string;
  role?: string;
}

function assertContractPayload(messageName: string, payload: Record<string, unknown>): void {
  const issues = validateContractPayload(messageName, payload);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
}

export const useLogin = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation({
    mutationFn: (payload: LoginPayload): Promise<AuthResponse> => {
      assertContractPayload('common.auth.v1.LoginRequest', {
        ...(payload as unknown as Record<string, unknown>),
      });
      return unwrap(safe.post<AuthResponse>('/auth/login', payload));
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
      login(data.role ?? '');
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
    mutationFn: (payload: RegisterPayload): Promise<AuthResponse> => {
      assertContractPayload('common.auth.v1.RegisterRequest', {
        ...(payload as unknown as Record<string, unknown>),
      });
      return unwrap(safe.post<AuthResponse>('/auth/register', payload));
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
      login(data.role ?? '');
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

export const checkEmailExists = (email: string): Promise<{ exists: boolean }> =>
  unwrap(safe.get<{ exists: boolean }>('/get-email-exists', { params: { email } }));

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
      login(data.role ?? '');
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
