import { todayIso } from './dateUtils';

export type DocumentStatus = 'Draft' | 'Pending Review' | 'Approved' | 'Rejected' | 'Expired';
export type DocumentCategory = 'Contract' | 'Policy' | 'Certificate' | 'ID' | 'Payslip' | 'Other';

export function documentStatusLabel(status: DocumentStatus): string {
  const labels: Record<DocumentStatus, string> = {
    Draft: 'Draft',
    'Pending Review': 'Pending Review',
    Approved: 'Approved',
    Rejected: 'Rejected',
    Expired: 'Expired',
  };
  return labels[status];
}

export function documentStatusColor(status: DocumentStatus): string {
  const colors: Record<DocumentStatus, string> = {
    Draft: 'gray',
    'Pending Review': 'yellow',
    Approved: 'green',
    Rejected: 'red',
    Expired: 'orange',
  };
  return colors[status];
}

export function isDocumentActive(status: DocumentStatus): boolean {
  return status === 'Approved';
}

export function requiresAction(status: DocumentStatus): boolean {
  return status === 'Draft' || status === 'Pending Review';
}

export function fileExtension(filename: string): string {
  const parts = filename.split('.');
  if (parts.length < 2) return '';
  return parts[parts.length - 1].toLowerCase();
}

export function isImageFile(filename: string): boolean {
  const ext = fileExtension(filename);
  return ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'].includes(ext);
}

export function isDocumentFile(filename: string): boolean {
  const ext = fileExtension(filename);
  return ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'txt', 'csv'].includes(ext);
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/**
 * Whether a document's expiry has already passed.
 *
 * A document expiring **today** is not expired: it is valid for the whole of
 * today. Comparing a parsed date to the current instant said otherwise from
 * 00:00 UTC, so a valid document was flagged red for its entire last day
 * (rtylr#1077).
 */
export function isExpired(expiryDateIso: string): boolean {
  return expiryDateIso.split('T')[0] < todayIso();
}

export function countByCategory(
  documents: Array<{ category: DocumentCategory }>
): Record<string, number> {
  const result: Record<string, number> = {};
  for (const doc of documents) {
    result[doc.category] = (result[doc.category] ?? 0) + 1;
  }
  return result;
}
