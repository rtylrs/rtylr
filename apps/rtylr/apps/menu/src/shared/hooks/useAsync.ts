import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { debounce } from 'lodash';

type AsyncFunction<T> = (data: unknown) => Promise<T>;
type SuccessCallbackFunction<T> = (result: T) => void;
type ErrorCallbackFunction = (message: string, error: unknown) => void;
type ErrorMessageFunction = (error: unknown) => string;

interface UseAsyncParams<T> {
  fn: AsyncFunction<T>;
  onSuccess?: SuccessCallbackFunction<T>;
  onError?: ErrorCallbackFunction;
  callOnMount?: boolean;
  callOnMountData?: unknown;
  debounceTime?: number;
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
}: UseAsyncParams<T>): UseAsyncReturn<T> {
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<boolean>(false);
  const [result, setResult] = useState<T | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [requestData, setRequestData] = useState<unknown>(null);

  const onSuccessRef = useRef(onSuccess);
  const onErrorRef = useRef(onError);

  useEffect(() => {
    onSuccessRef.current = onSuccess;
    onErrorRef.current = onError;
  }, [onSuccess, onError]);

  const execute = useCallback(
    async (data: unknown) => {
      try {
        setError(null);
        setSuccess(false);
        setLoading(true);
        setRequestData(data);
        const res = await fn(data);
        setResult(res);
        setSuccess(true);
        onSuccessRef.current?.(res);
      } catch (error) {
        const message = errorMessage(error);
        setError(message);
        onErrorRef.current?.(message, error);
      } finally {
        setLoading(false);
      }
    },
    [fn]
  );

  const debouncedMain = useMemo(() => debounce(execute, debounceTime), [debounceTime, execute]);

  useEffect(() => {
    if (callOnMount) {
      debouncedMain(callOnMountData);
    }
  }, [callOnMount, callOnMountData, debouncedMain]);

  useEffect(() => {
    return () => {
      debouncedMain.cancel();
    };
  }, [debouncedMain]);

  return { main: debouncedMain, error, success, loading, result, requestData };
}
