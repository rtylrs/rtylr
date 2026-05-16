/** Race a promise against a timeout; rejects with a TimeoutError on expiry. */
export function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const id = setTimeout(() => {
      const e = new Error(`Timed out after ${ms}ms`);
      e.name = 'TimeoutError';
      reject(e);
    }, ms);
    promise.then(
      (v) => {
        clearTimeout(id);
        resolve(v);
      },
      (err) => {
        clearTimeout(id);
        reject(err);
      }
    );
  });
}
