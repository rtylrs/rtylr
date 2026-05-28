import { describe, it, expect } from 'vitest';

import {
  commentsForTask,
  sortCommentsAsc,
  sortCommentsDesc,
  commentCount,
  commentMentions,
  commentsWithMention,
  latestComment,
  commentsByAuthor,
} from './commentUtils';
import type { Comment } from '../stores/flowStore';

function makeComment(overrides: Partial<Comment>): Comment {
  return {
    id: 'c1',
    taskId: 't1',
    authorId: 'u1',
    author: 'Alice',
    authorInitials: 'AL',
    body: 'Hello',
    createdAt: '2026-05-01T10:00:00Z',
    mentions: [],
    ...overrides,
  } as Comment;
}

describe('commentsForTask', () => {
  it('returns comments for the given task', () => {
    const comments = [
      makeComment({ id: 'a', taskId: 't1' }),
      makeComment({ id: 'b', taskId: 't2' }),
    ];
    expect(commentsForTask(comments, 't1').map((c) => c.id)).toEqual(['a']);
  });
});

describe('sortCommentsAsc', () => {
  it('sorts oldest first', () => {
    const comments = [
      makeComment({ id: 'a', createdAt: '2026-06-01T00:00:00Z' }),
      makeComment({ id: 'b', createdAt: '2026-01-01T00:00:00Z' }),
    ];
    expect(sortCommentsAsc(comments).map((c) => c.id)).toEqual(['b', 'a']);
  });

  it('does not mutate original', () => {
    const comments = [makeComment({ createdAt: '2026-06-01T00:00:00Z' }), makeComment()];
    const original = [...comments];
    sortCommentsAsc(comments);
    expect(comments).toEqual(original);
  });
});

describe('sortCommentsDesc', () => {
  it('sorts newest first', () => {
    const comments = [
      makeComment({ id: 'a', createdAt: '2026-01-01T00:00:00Z' }),
      makeComment({ id: 'b', createdAt: '2026-06-01T00:00:00Z' }),
    ];
    expect(sortCommentsDesc(comments).map((c) => c.id)).toEqual(['b', 'a']);
  });
});

describe('commentCount', () => {
  it('counts comments for a task', () => {
    const comments = [
      makeComment({ taskId: 't1' }),
      makeComment({ taskId: 't1' }),
      makeComment({ taskId: 't2' }),
    ];
    expect(commentCount(comments, 't1')).toBe(2);
  });
});

describe('commentMentions', () => {
  it('returns true when user is mentioned', () => {
    const c = makeComment({ mentions: ['u1', 'u2'] });
    expect(commentMentions(c, 'u1')).toBe(true);
  });

  it('returns false when user is not mentioned', () => {
    const c = makeComment({ mentions: ['u2'] });
    expect(commentMentions(c, 'u1')).toBe(false);
  });
});

describe('commentsWithMention', () => {
  it('returns comments mentioning the user', () => {
    const comments = [
      makeComment({ id: 'a', mentions: ['u1'] }),
      makeComment({ id: 'b', mentions: ['u2'] }),
    ];
    expect(commentsWithMention(comments, 'u1').map((c) => c.id)).toEqual(['a']);
  });
});

describe('latestComment', () => {
  it('returns the most recent comment for a task', () => {
    const comments = [
      makeComment({ id: 'a', taskId: 't1', createdAt: '2026-01-01T00:00:00Z' }),
      makeComment({ id: 'b', taskId: 't1', createdAt: '2026-06-01T00:00:00Z' }),
    ];
    expect(latestComment(comments, 't1')?.id).toBe('b');
  });

  it('returns undefined for unknown task', () => {
    expect(latestComment([], 'unknown')).toBeUndefined();
  });
});

describe('commentsByAuthor', () => {
  it('returns comments by a specific author', () => {
    const comments = [
      makeComment({ id: 'a', authorId: 'u1' }),
      makeComment({ id: 'b', authorId: 'u2' }),
      makeComment({ id: 'c', authorId: 'u1' }),
    ];
    expect(commentsByAuthor(comments, 'u1').map((c) => c.id)).toEqual(['a', 'c']);
  });
});
