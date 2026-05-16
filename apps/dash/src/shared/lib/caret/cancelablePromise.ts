export interface CancelablePromise<T> {
  promise: Promise<T>;
  cancel: () => void;
}

/** Wrap a promise with a `cancel()` that rejects the wrapped promise. */
export function makeCancelable<T>(promise: Promise<T>): CancelablePromise<T> {
  let cancelled = false;
  const wrapped = new Promise<T>((resolve, reject) => {
    promise.then(
      (v) => (cancelled ? reject(new Error('cancelled')) : resolve(v)),
      (err) => reject(cancelled ? new Error('cancelled') : err)
    );
  });
  return {
    promise: wrapped,
    cancel: (): void => {
      cancelled = true;
    },
  };
}
