import { useCallback, useEffect, useState } from "react";
import { debounce } from "lodash"; // Import debounce from lodash

const errorMessage = (error) => {
  return (
    (error.response && error.response.data && error.response.data.message) ||
    error.message ||
    error.toString()
  );
};

export default function useAsync({
  fn = () => { },
  onSuccess = () => { },
  onError = () => { },
  callOnMount = false,
  callOnMountData = null,
  debounceTime = 0,
  dependencies = [],
}) {
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [requestData, setRequestData] = useState(null);

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
