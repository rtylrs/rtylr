/**
 * Text sanitization and mojibake cleaning for Recruit copy.
 */
export function cleanMojibake(text: string | null | undefined): string {
  if (!text) return '';
  return text
    .replace(/â€”/g, '—')
    .replace(/â€“/g, '–')
    .replace(/â€™/g, '’')
    .replace(/â€˜/g, '‘')
    .replace(/â€œ/g, '“')
    .replace(/â€/g, '”')
    .replace(/â€¦/g, '…')
    .replace(/Â /g, ' ')
    .replace(/Â/g, '')
    .trim();
}
