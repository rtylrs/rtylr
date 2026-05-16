/** Friendly error copy for the in-chat error block. */
export function friendlyError(error: unknown): string {
  if (!error) return 'Something went wrong.';
  if (error instanceof Error) {
    if (error.message.includes('Failed to fetch')) {
      return 'Lost connection. Check your internet and try again.';
    }
    if (error.message.includes('429')) {
      return 'Too many requests. Take a breath and try again in a minute.';
    }
    if (error.message.includes('401')) {
      return 'Your session expired. Please refresh the page.';
    }
    if (error.message.includes('500')) {
      return 'The server had a hiccup. Try again in a moment.';
    }
    return error.message;
  }
  return 'Something went wrong.';
}
