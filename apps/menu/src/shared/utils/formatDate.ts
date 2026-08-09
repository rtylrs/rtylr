export function formatDate(dateString) {
  const date = new Date(dateString);
  if (isNaN(date)) {
    return 'Invalid Date';
  }
  return date.toLocaleDateString();
}
