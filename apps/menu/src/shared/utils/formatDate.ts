export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) {
    return 'Invalid Date';
  }
  return date.toLocaleDateString();
}
