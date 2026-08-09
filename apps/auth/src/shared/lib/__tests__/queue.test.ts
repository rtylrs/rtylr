import { describe, it, expect } from 'vitest';
import { Queue, runSerial, runBatched } from '../queue';

describe('Queue', () => {
  it('enqueues and dequeues in FIFO order', () => {
    const q = new Queue<number>();
    q.enqueue(1);
    q.enqueue(2);
    expect(q.dequeue()).toBe(1);
    expect(q.dequeue()).toBe(2);
  });

  it('peek returns front without removing', () => {
    const q = new Queue<string>();
    q.enqueue('a');
    expect(q.peek()).toBe('a');
    expect(q.size).toBe(1);
  });

  it('isEmpty returns true for empty queue', () => {
    expect(new Queue().isEmpty()).toBe(true);
  });

  it('clear empties the queue', () => {
    const q = new Queue<number>();
    q.enqueue(1);
    q.clear();
    expect(q.isEmpty()).toBe(true);
  });

  it('toArray returns a copy', () => {
    const q = new Queue<number>();
    q.enqueue(1);
    q.enqueue(2);
    const arr = q.toArray();
    expect(arr).toEqual([1, 2]);
    arr.push(3);
    expect(q.size).toBe(2);
  });

  it('dequeue on empty returns undefined', () => {
    expect(new Queue().dequeue()).toBeUndefined();
  });
});

describe('runSerial', () => {
  it('runs tasks in order', async () => {
    const order: number[] = [];
    await runSerial([
      async () => {
        order.push(1);
      },
      async () => {
        order.push(2);
      },
    ]);
    expect(order).toEqual([1, 2]);
  });

  it('returns all results', async () => {
    const results = await runSerial([async () => 'a', async () => 'b']);
    expect(results).toEqual(['a', 'b']);
  });
});

describe('runBatched', () => {
  it('returns results in original order', async () => {
    const tasks = [1, 2, 3, 4].map((n) => async () => n * 2);
    const results = await runBatched(tasks, 2);
    expect(results).toEqual([2, 4, 6, 8]);
  });

  it('handles concurrency of 1 (serial)', async () => {
    const tasks = [async () => 1, async () => 2, async () => 3];
    expect(await runBatched(tasks, 1)).toEqual([1, 2, 3]);
  });
});
