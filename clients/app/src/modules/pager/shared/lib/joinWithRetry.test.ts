import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  JOIN_RETRY_INTERVAL_MS,
  JOIN_RETRY_WINDOW_MS,
  isAbort,
  joinWithRetry,
} from './joinWithRetry';

const notYet = { statusCode: 404, code: 'ticket_not_found' };
const isNotYet = (error: unknown): boolean => error === notYet;

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('joinWithRetry', () => {
  it('returns the first answer without waiting when the number is known', async () => {
    const join = vi.fn(async () => 'token');
    const onWaiting = vi.fn();

    await expect(joinWithRetry({ join, isNotYet, onWaiting })).resolves.toBe('token');
    expect(join).toHaveBeenCalledTimes(1);
    expect(onWaiting).not.toHaveBeenCalled();
  });

  it('asks again every five seconds until the cashier has entered the number', async () => {
    const join = vi
      .fn<() => Promise<string>>()
      .mockRejectedValueOnce(notYet)
      .mockRejectedValueOnce(notYet)
      .mockResolvedValue('token');
    const onWaiting = vi.fn();

    const result = joinWithRetry({ join, isNotYet, onWaiting });
    await vi.advanceTimersByTimeAsync(0);
    expect(join).toHaveBeenCalledTimes(1);
    expect(onWaiting).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(JOIN_RETRY_INTERVAL_MS - 1);
    expect(join).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(1);
    expect(join).toHaveBeenCalledTimes(2);

    await vi.advanceTimersByTimeAsync(JOIN_RETRY_INTERVAL_MS);
    await expect(result).resolves.toBe('token');
    expect(onWaiting).toHaveBeenCalledTimes(1);
  });

  it('gives up after two minutes and hands back the last answer', async () => {
    const join = vi.fn(async (): Promise<string> => {
      throw notYet;
    });

    const result = joinWithRetry({ join, isNotYet });
    const settled = expect(result).rejects.toBe(notYet);
    await vi.advanceTimersByTimeAsync(JOIN_RETRY_WINDOW_MS);
    await settled;

    expect(join).toHaveBeenCalledTimes(JOIN_RETRY_WINDOW_MS / JOIN_RETRY_INTERVAL_MS + 1);

    await vi.advanceTimersByTimeAsync(JOIN_RETRY_INTERVAL_MS * 2);
    expect(join).toHaveBeenCalledTimes(JOIN_RETRY_WINDOW_MS / JOIN_RETRY_INTERVAL_MS + 1);
  });

  it('does not retry any other failure', async () => {
    const failure = { statusCode: 400, code: 'invalid_number' };
    const join = vi.fn(async (): Promise<string> => {
      throw failure;
    });

    await expect(joinWithRetry({ join, isNotYet })).rejects.toBe(failure);
    expect(join).toHaveBeenCalledTimes(1);
  });

  it('stops asking when the customer cancels', async () => {
    const join = vi.fn(async (): Promise<string> => {
      throw notYet;
    });
    const controller = new AbortController();

    const result = joinWithRetry({ join, isNotYet, signal: controller.signal });
    const settled = expect(result).rejects.toSatisfy(isAbort);
    await vi.advanceTimersByTimeAsync(JOIN_RETRY_INTERVAL_MS);
    expect(join).toHaveBeenCalledTimes(2);

    controller.abort();
    await settled;
    await vi.advanceTimersByTimeAsync(JOIN_RETRY_INTERVAL_MS * 3);

    expect(join).toHaveBeenCalledTimes(2);
  });
});
