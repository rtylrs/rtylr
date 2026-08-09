// Candidate document management helpers for the recruit admin panel.

export type DocumentType = 'resume' | 'cover_letter' | 'portfolio' | 'reference' | 'other';
export type DocumentStatus = 'pending' | 'approved' | 'rejected';

export interface CandidateDocument {
  id: string;
  candidateId: string;
  type: DocumentType;
  status: DocumentStatus;
  filename: string;
  uploadedAt: string;
  sizeBytes: number;
}

/**
 * documentTypeLabel returns a display-friendly label for a document type.
 */
export function documentTypeLabel(type: DocumentType | string): string {
  const labels: Record<string, string> = {
    resume: 'Resume',
    cover_letter: 'Cover Letter',
    portfolio: 'Portfolio',
    reference: 'Reference Letter',
    other: 'Other',
  };
  return labels[type] ?? type;
}

/**
 * documentsByType groups documents by their type.
 */
export function documentsByType(docs: CandidateDocument[]): Record<string, CandidateDocument[]> {
  return docs.reduce<Record<string, CandidateDocument[]>>((acc, d) => {
    acc[d.type] = [...(acc[d.type] ?? []), d];
    return acc;
  }, {});
}

/**
 * approvedDocuments returns documents with approved status.
 */
export function approvedDocuments(docs: CandidateDocument[]): CandidateDocument[] {
  return docs.filter((d) => d.status === 'approved');
}

/**
 * hasRequiredDocument returns true when the candidate has a document of the given type.
 */
export function hasRequiredDocument(docs: CandidateDocument[], type: DocumentType): boolean {
  return docs.some((d) => d.type === type);
}

/**
 * totalDocumentSize returns the total size in bytes of all documents.
 */
export function totalDocumentSize(docs: CandidateDocument[]): number {
  return docs.reduce((sum, d) => sum + d.sizeBytes, 0);
}

/**
 * formatDocumentSize formats bytes as KB or MB.
 */
export function formatDocumentSize(bytes: number): string {
  if (bytes >= 1024 * 1024) {
    return `${Math.round(bytes / (1024 * 1024))} MB`;
  }
  return `${Math.round(bytes / 1024)} KB`;
}
