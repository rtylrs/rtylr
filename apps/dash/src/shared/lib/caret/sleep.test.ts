import { describe, expect, it } from 'vitest';
import { sleep } from './sleep';

describe('sleep', () => {
  it('resolves after the requested duration', async () => {
    const before = Date.now();
    await sleep(20);
    expect(Date.now() - before).toBeGreaterThanOrEqual(15);
  });
});
