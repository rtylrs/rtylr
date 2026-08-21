import { useMutation } from '@tanstack/react-query';
import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { AuthServiceClient } from '@voxire/voxie/common/auth/v1/client';
import type {
  GoogleAuthRequest,
  GoogleAuthResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  CheckEmailExistsResponse,
} from '@voxire/voxie/common/auth/v1/types';
import { validateContractPayload } from '@voxire/voxie/validation/metadata';
import { useUserStore } from '../stores/userStore';
import { setCookie, getDefaultCookieDomain } from '@voxire/auth';

const authClient = createOneKitClient(AuthServiceClient, 'auth');

function assertContractPayload(messageName: string, payload: Record<string, unknown>): void {
  const issues = validateContractPayload(messageName, payload);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
}

function persistSession(login: (role: string) => void, token: string): void {
  setCookie('token', token, {
    days: 30,
    sameSite: 'lax',
    domain: getDefaultCookieDomain(),
  });
  login('');
}

export const useLogin = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation<LoginResponse, Error, LoginRequest>({
    mutationFn: (payload) => {
      assertContractPayload('common.auth.v1.LoginRequest', {
        ...(payload as unknown as Record<string, unknown>),
      });
      return authClient.login(payload);
    },
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data) => persistSession(login, data.token),
    onError: (error) => {
      setError(getOneKitErrorMessage(error, 'Login failed'));
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const useRegisterRestaurant = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation<RegisterResponse, Error, RegisterRequest>({
    mutationFn: (payload) => {
      assertContractPayload('common.auth.v1.RegisterRequest', {
        ...(payload as unknown as Record<string, unknown>),
      });
      return authClient.register(payload);
    },
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data, variables) => {
      if (variables.businessName) {
        try {
          localStorage.setItem('rtylr_business_name', variables.businessName);
        } catch {
          // ignore storage quota errors
        }
      }
      persistSession(login, data.token);
    },
    onError: (error) => {
      setError(getOneKitErrorMessage(error, 'Registration failed'));
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};

export const checkEmailExists = (email: string): Promise<CheckEmailExistsResponse> =>
  authClient.checkEmailExists({ email });

export const useGoogleAuth = () => {
  const { login, setLoading, setError } = useUserStore();

  return useMutation<GoogleAuthResponse, Error, GoogleAuthRequest>({
    mutationFn: (payload) => authClient.googleAuth(payload),
    onMutate: () => {
      setLoading(true);
      setError(null);
    },
    onSuccess: (data) => persistSession(login, data.token),
    onError: (error) => {
      setError(getOneKitErrorMessage(error, 'Google login failed'));
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};
