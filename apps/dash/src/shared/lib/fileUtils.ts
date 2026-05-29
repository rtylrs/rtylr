/**
 * File and upload helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type FileCategory = 'image' | 'document' | 'spreadsheet' | 'video' | 'audio' | 'archive' | 'other';

const MIME_CATEGORIES: Record<string, FileCategory> = {
  'image/jpeg': 'image',
  'image/png': 'image',
  'image/gif': 'image',
  'image/webp': 'image',
  'image/svg+xml': 'image',
  'application/pdf': 'document',
  'application/msword': 'document',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'document',
  'application/vnd.ms-excel': 'spreadsheet',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'spreadsheet',
  'text/csv': 'spreadsheet',
  'video/mp4': 'video',
  'video/webm': 'video',
  'audio/mpeg': 'audio',
  'audio/wav': 'audio',
  'application/zip': 'archive',
  'application/x-tar': 'archive',
  'application/gzip': 'archive',
};

/** Returns the file category for a given MIME type. */
export function fileCategoryFromMime(mimeType: string): FileCategory {
  return MIME_CATEGORIES[mimeType] ?? 'other';
}

/** Returns a human-readable label for a file category. */
export function fileCategoryLabel(category: FileCategory): string {
  const labels: Record<FileCategory, string> = {
    image: 'Image',
    document: 'Document',
    spreadsheet: 'Spreadsheet',
    video: 'Video',
    audio: 'Audio',
    archive: 'Archive',
    other: 'File',
  };
  return labels[category];
}

/**
 * Formats a file size in bytes to a human-readable string.
 * e.g. 1024 -> '1.0 KB', 1048576 -> '1.0 MB'
 */
export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}

/** Returns the file extension from a filename (lowercase, without dot). */
export function fileExtension(filename: string): string {
  const idx = filename.lastIndexOf('.');
  return idx === -1 ? '' : filename.slice(idx + 1).toLowerCase();
}

/** Returns true if the file extension is in the allowed list. */
export function isAllowedExtension(filename: string, allowed: string[]): boolean {
  const ext = fileExtension(filename);
  return allowed.map((a) => a.toLowerCase().replace(/^\./, '')).includes(ext);
}

/** Returns true if the file size is within the max bytes limit. */
export function isWithinSizeLimit(bytes: number, maxBytes: number): boolean {
  return bytes <= maxBytes;
}

/** Strips potentially unsafe characters from a filename for safe storage. */
export function sanitizeFilename(filename: string): string {
  return filename
    .replace(/[<>:"\/\\|?*]/g, '_')
    .replace(/\s+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '');
}
