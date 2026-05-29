import type { Comment } from '../stores/flowStore';

/** Returns all comments for the given task ID. */
export function commentsForTask(comments: Comment[], taskId: string): Comment[] {
  return comments.filter((c) => c.taskId === taskId);
}

/** Sorts comments oldest-first (ascending by createdAt). Does not mutate. */
export function sortCommentsAsc(comments: Comment[]): Comment[] {
  return [...comments].sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
  );
}

/** Sorts comments newest-first (descending by createdAt). Does not mutate. */
export function sortCommentsDesc(comments: Comment[]): Comment[] {
  return [...comments].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Returns the count of comments for a given task ID. */
export function commentCount(comments: Comment[], taskId: string): number {
  return comments.filter((c) => c.taskId === taskId).length;
}

/** Returns true when the comment's mentions array includes the userId. */
export function commentMentions(comment: Comment, userId: string): boolean {
  return (comment.mentions ?? []).includes(userId);
}

/** Returns comments that mention the given userId. */
export function commentsWithMention(comments: Comment[], userId: string): Comment[] {
  return comments.filter((c) => commentMentions(c, userId));
}

/** Returns the most recent comment for a task, or undefined if none. */
export function latestComment(comments: Comment[], taskId: string): Comment | undefined {
  const taskComments = commentsForTask(comments, taskId);
  if (taskComments.length === 0) return undefined;
  return sortCommentsDesc(taskComments)[0];
}

/** Returns all comments authored by the given user. */
export function commentsByAuthor(comments: Comment[], authorId: string): Comment[] {
  return comments.filter((c) => c.authorId === authorId);
}
