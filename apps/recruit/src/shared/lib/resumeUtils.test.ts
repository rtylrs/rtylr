import { describe, it, expect } from 'vitest';
import {
  isAcceptedMimeType,
  isAcceptedExtension,
  isFileSizeValid,
  formatFileSize,
  acceptedFormats,
} from './resumeUtils';

describe('isAcceptedMimeType', () => {
  it('accepts PDF MIME type', () => {
    expect(isAcceptedMimeType('application/pdf')).toBe(true);
  });

  it('accepts Word document MIME types', () => {
    expect(isAcceptedMimeType('application/msword')).toBe(true);
    expect(
      isAcceptedMimeType('application/vnd.openxmlformats-officedocument.wordprocessingml.document')
    ).toBe(true);
  });

  it('rejects image MIME types', () => {
    expect(isAcceptedMimeType('image/jpeg')).toBe(false);
    expect(isAcceptedMimeType('text/plain')).toBe(false);
  });
});

describe('isAcceptedExtension', () => {
  it('accepts .pdf, .doc, .docx', () => {
    expect(isAcceptedExtension('resume.pdf')).toBe(true);
    expect(isAcceptedExtension('resume.doc')).toBe(true);
    expect(isAcceptedExtension('resume.docx')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(isAcceptedExtension('RESUME.PDF')).toBe(true);
  });

  it('rejects .txt, .png', () => {
    expect(isAcceptedExtension('resume.txt')).toBe(false);
    expect(isAcceptedExtension('photo.png')).toBe(false);
  });
});

describe('isFileSizeValid', () => {
  it('accepts files under 5 MB', () => {
    expect(isFileSizeValid(1024 * 1024)).toBe(true);
  });

  it('accepts exactly 5 MB', () => {
    expect(isFileSizeValid(5 * 1024 * 1024)).toBe(true);
  });

  it('rejects files over 5 MB', () => {
    expect(isFileSizeValid(5 * 1024 * 1024 + 1)).toBe(false);
  });
});

describe('formatFileSize', () => {
  it('formats bytes', () => {
    expect(formatFileSize(500)).toBe('500 B');
  });

  it('formats kilobytes', () => {
    expect(formatFileSize(1536)).toBe('1.5 KB');
  });

  it('formats megabytes', () => {
    expect(formatFileSize(2 * 1024 * 1024)).toBe('2.0 MB');
  });
});

describe('acceptedFormats', () => {
  it('returns a descriptive string', () => {
    expect(acceptedFormats()).toContain('PDF');
    expect(acceptedFormats()).toContain('5 MB');
  });
});
