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

/**
 * Store the session cookie, and report whether there was a session to store.
 *
 * The token is optional on every generated response type, so a caller cannot
 * assume it arrived. Persisting `undefined` would write a cookie that reads
 * back as the string "undefined" and leave the user signed in to nothing, so
 * the absence is returned and each caller surfaces its own error instead.
 */
function persistSession(login: () => void, token: string | undefined): boolean {
  if (token === undefined || token === '') return false;
  setCookie('token', token, {
    days: 30,
    sameSite: 'lax',
    domain: getDefaultCookieDomain(),
  });
  login();
  return true;
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
    onSuccess: (data) => {
      if (!persistSession(login, data.token)) setError('Login failed');
    },
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
    onSuccess: (data) => {
      // The business name is not stashed here any more. It was written to
      // localStorage['rtylr_business_name'] and never read back, by anything,
      // in any app or package (#1036, F-1).
      //
      // It is not lost: `business_name` is a real field on
      // common.auth.v1.RegisterRequest, the generated encoder maps it, and the
      // register call already carries it. Keeping a second copy in storage
      // only suggested the request did not.
      if (!persistSession(login, data.token)) setError('Registration failed');
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
    onSuccess: (data) => {
      if (!persistSession(login, data.token)) setError('Google login failed');
    },
    onError: (error) => {
      setError(getOneKitErrorMessage(error, 'Google login failed'));
      setLoading(false);
    },
    onSettled: () => {
      setLoading(false);
    },
  });
};
