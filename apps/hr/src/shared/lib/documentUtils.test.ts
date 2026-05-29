import { describe, expect, it } from 'vitest';
import {
  countByCategory,
  documentStatusColor,
  documentStatusLabel,
  fileExtension,
  formatFileSize,
  isDocumentActive,
  isDocumentFile,
  isImageFile,
  requiresAction,
} from './documentUtils';

describe('documentStatusLabel', () => {
  it('returns correct label for each status', () => {
    expect(documentStatusLabel('Draft')).toBe('Draft');
    expect(documentStatusLabel('Pending Review')).toBe('Pending Review');
    expect(documentStatusLabel('Approved')).toBe('Approved');
    expect(documentStatusLabel('Rejected')).toBe('Rejected');
    expect(documentStatusLabel('Expired')).toBe('Expired');
  });
});

describe('documentStatusColor', () => {
  it('returns correct color for each status', () => {
    expect(documentStatusColor('Draft')).toBe('gray');
    expect(documentStatusColor('Pending Review')).toBe('yellow');
    expect(documentStatusColor('Approved')).toBe('green');
    expect(documentStatusColor('Rejected')).toBe('red');
    expect(documentStatusColor('Expired')).toBe('orange');
  });
});

describe('isDocumentActive', () => {
  it('returns true only for Approved', () => {
    expect(isDocumentActive('Approved')).toBe(true);
    expect(isDocumentActive('Draft')).toBe(false);
    expect(isDocumentActive('Pending Review')).toBe(false);
    expect(isDocumentActive('Rejected')).toBe(false);
    expect(isDocumentActive('Expired')).toBe(false);
  });
});

describe('requiresAction', () => {
  it('returns true for Draft and Pending Review', () => {
    expect(requiresAction('Draft')).toBe(true);
    expect(requiresAction('Pending Review')).toBe(true);
    expect(requiresAction('Approved')).toBe(false);
    expect(requiresAction('Rejected')).toBe(false);
    expect(requiresAction('Expired')).toBe(false);
  });
});

describe('fileExtension', () => {
  it('extracts lowercase extension', () => {
    expect(fileExtension('document.PDF')).toBe('pdf');
    expect(fileExtension('report.xlsx')).toBe('xlsx');
    expect(fileExtension('image.JPG')).toBe('jpg');
  });

  it('returns empty string if no extension', () => {
    expect(fileExtension('nodotfile')).toBe('');
  });
});

describe('isImageFile', () => {
  it('returns true for image files', () => {
    expect(isImageFile('photo.jpg')).toBe(true);
    expect(isImageFile('logo.PNG')).toBe(true);
    expect(isImageFile('icon.svg')).toBe(true);
  });

  it('returns false for non-image files', () => {
    expect(isImageFile('contract.pdf')).toBe(false);
    expect(isImageFile('data.xlsx')).toBe(false);
  });
});

describe('isDocumentFile', () => {
  it('returns true for document files', () => {
    expect(isDocumentFile('report.pdf')).toBe(true);
    expect(isDocumentFile('data.xlsx')).toBe(true);
    expect(isDocumentFile('notes.txt')).toBe(true);
  });

  it('returns false for non-document files', () => {
    expect(isDocumentFile('photo.jpg')).toBe(false);
  });
});

describe('formatFileSize', () => {
  it('formats bytes correctly', () => {
    expect(formatFileSize(500)).toBe('500 B');
    expect(formatFileSize(1024)).toBe('1.0 KB');
    expect(formatFileSize(1048576)).toBe('1.0 MB');
    expect(formatFileSize(2097152)).toBe('2.0 MB');
  });
});

describe('countByCategory', () => {
  it('counts documents by category', () => {
    const docs = [
      { category: 'Contract' as const },
      { category: 'Policy' as const },
      { category: 'Contract' as const },
      { category: 'ID' as const },
    ];
    const result = countByCategory(docs);
    expect(result['Contract']).toBe(2);
    expect(result['Policy']).toBe(1);
    expect(result['ID']).toBe(1);
    expect(result['Payslip']).toBeUndefined();
  });

  it('returns empty object for empty array', () => {
    expect(countByCategory([])).toEqual({});
  });
});
