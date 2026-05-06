import { useCallback, useEffect, useState } from 'react';
import { debounce } from 'lodash';

type AsyncFunction<T> = (data: unknown) => Promise<T>;
type CallbackFunction = () => void;
type ErrorMessageFunction = (error: unknown) => string;

interface UseAsyncParams<T> {
  fn: AsyncFunction<T>;
  onSuccess?: CallbackFunction;
  onError?: CallbackFunction;
  callOnMount?: boolean;
  callOnMountData?: unknown;
  debounceTime?: number;
  dependencies?: unknown[];
}

interface UseAsyncReturn<T> {
  main: (data: unknown) => void;
  error: string | null;
  success: boolean;
  loading: boolean;
  result: T | null;
  requestData: unknown;
}

const errorMessage: ErrorMessageFunction = (error) => {
  const err = error as { response?: { data?: { message?: string } }; message?: string };
  return err?.response?.data?.message || err?.message || String(error);
};

export default function useAsync<T>({
  fn,
  onSuccess = () => {},
  onError = () => {},
  callOnMount = false,
  callOnMountData = null,
  debounceTime = 0,
  dependencies = [],
}: UseAsyncParams<T>): UseAsyncReturn<T> {
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);
  const [result, setResult] = useState<T | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [requestData, setRequestData] = useState<unknown>(null);

  const debouncedMain = useCallback(
    debounce(async (data) => {
      try {
        setLoading(true);
        setRequestData(data);
        const res = await fn(data);
        setResult(res);
        setSuccess(true);
      } catch (error) {
        setError(errorMessage(error));
      } finally {
        setLoading(false);
      }
    }, debounceTime),
    [fn, debounceTime]
  );

  function reset() {
    setError(null);
    setSuccess(false);
    setResult(null);
    setLoading(false);
  }

  useEffect(() => {
    if (callOnMount) {
      debouncedMain(callOnMountData);
    }
  }, [...dependencies]);

  useEffect(() => {
    if (success) {
      onSuccess();
    }

    if (error) {
      onError();
    }

    return () => {
      reset();
    };
  }, [success, error]);

  return { main: debouncedMain, error, success, loading, result, requestData };
}
