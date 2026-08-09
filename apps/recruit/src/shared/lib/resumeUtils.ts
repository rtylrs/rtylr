// Resume file helpers for the recruit apply form.

const ACCEPTED_MIME_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
];

const ACCEPTED_EXTENSIONS = ['.pdf', '.doc', '.docx'];

const MAX_RESUME_BYTES = 5 * 1024 * 1024; // 5 MB

/**
 * isAcceptedMimeType returns true when the MIME type is allowed for resumes.
 */
export function isAcceptedMimeType(mimeType: string): boolean {
  return ACCEPTED_MIME_TYPES.includes(mimeType.toLowerCase());
}

/**
 * isAcceptedExtension returns true when the file extension is allowed.
 */
export function isAcceptedExtension(filename: string): boolean {
  const lower = filename.toLowerCase();
  return ACCEPTED_EXTENSIONS.some((ext) => lower.endsWith(ext));
}

/**
 * isFileSizeValid returns true when the file size is within the 5 MB limit.
 */
export function isFileSizeValid(bytes: number): boolean {
  return bytes <= MAX_RESUME_BYTES;
}

/**
 * formatFileSize returns a human-readable file size string.
 */
export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/**
 * acceptedFormats returns a string describing the accepted file formats.
 */
export function acceptedFormats(): string {
  return 'PDF, DOC, DOCX (max 5 MB)';
}
