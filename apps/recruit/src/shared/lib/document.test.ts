import { describe, it, expect } from 'vitest';
import {
  documentTypeLabel,
  documentsByType,
  approvedDocuments,
  hasRequiredDocument,
  totalDocumentSize,
  formatDocumentSize,
} from './document';
import type { CandidateDocument } from './document';

function makeDoc(
  type: 'resume' | 'cover_letter' | 'portfolio' | 'reference' | 'other',
  status: 'pending' | 'approved' | 'rejected' = 'pending',
  sizeBytes = 100000
): CandidateDocument {
  return {
    id: 'd1',
    candidateId: 'c1',
    type,
    status,
    filename: `doc.pdf`,
    uploadedAt: '2026-01-01',
    sizeBytes,
  };
}

const docs = [
  makeDoc('resume', 'approved', 500000),
  makeDoc('cover_letter', 'pending', 200000),
  makeDoc('portfolio', 'rejected', 1500000),
];

describe('documentTypeLabel', () => {
  it('returns display label for known types', () => {
    expect(documentTypeLabel('resume')).toBe('Resume');
    expect(documentTypeLabel('cover_letter')).toBe('Cover Letter');
  });

  it('returns raw value for unknown type', () => {
    expect(documentTypeLabel('transcript')).toBe('transcript');
  });
});

describe('documentsByType', () => {
  it('groups documents by type', () => {
    const grouped = documentsByType(docs);
    expect(grouped['resume']).toHaveLength(1);
    expect(grouped['portfolio']).toHaveLength(1);
  });
});

describe('approvedDocuments', () => {
  it('returns only approved documents', () => {
    expect(approvedDocuments(docs)).toHaveLength(1);
    expect(approvedDocuments(docs)[0].type).toBe('resume');
  });
});

describe('hasRequiredDocument', () => {
  it('returns true when document type exists', () => {
    expect(hasRequiredDocument(docs, 'resume')).toBe(true);
  });

  it('returns false when document type is missing', () => {
    expect(hasRequiredDocument(docs, 'reference')).toBe(false);
  });
});

describe('totalDocumentSize', () => {
  it('sums sizes of all documents', () => {
    expect(totalDocumentSize(docs)).toBe(2200000);
  });
});

describe('formatDocumentSize', () => {
  it('formats bytes as KB for small files', () => {
    expect(formatDocumentSize(204800)).toBe('200 KB');
  });

  it('formats bytes as MB for large files', () => {
    expect(formatDocumentSize(2097152)).toBe('2 MB');
  });
});
