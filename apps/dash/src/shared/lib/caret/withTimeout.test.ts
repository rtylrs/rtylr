import { describe, expect, it } from 'vitest';
import { withTimeout } from './withTimeout';
import { sleep } from './sleep';

describe('withTimeout', () => {
  it('resolves before the timeout', async () => {
    await expect(withTimeout(Promise.resolve(42), 100)).resolves.toBe(42);
  });
  it('rejects with TimeoutError when the timeout fires', async () => {
    await expect(
      withTimeout(
        sleep(50).then(() => 1),
        5
      )
    ).rejects.toMatchObject({
      name: 'TimeoutError',
    });
  });
});
