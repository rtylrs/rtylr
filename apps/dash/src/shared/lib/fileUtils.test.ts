import { describe, it, expect } from 'vitest';
import {
  fileCategoryFromMime,
  fileCategoryLabel,
  formatFileSize,
  fileExtension,
  isAllowedExtension,
  isWithinSizeLimit,
  sanitizeFilename,
} from './fileUtils';

describe('fileCategoryFromMime', () => {
  it('identifies image MIME types', () => expect(fileCategoryFromMime('image/png')).toBe('image'));
  it('identifies document MIME types', () => expect(fileCategoryFromMime('application/pdf')).toBe('document'));
  it('returns other for unknown MIME', () => expect(fileCategoryFromMime('application/octet-stream')).toBe('other'));
});

describe('fileCategoryLabel', () => {
  it('returns correct labels', () => {
    expect(fileCategoryLabel('image')).toBe('Image');
    expect(fileCategoryLabel('document')).toBe('Document');
    expect(fileCategoryLabel('other')).toBe('File');
  });
});

describe('formatFileSize', () => {
  it('bytes', () => expect(formatFileSize(500)).toBe('500 B'));
  it('kilobytes', () => expect(formatFileSize(2048)).toBe('2.0 KB'));
  it('megabytes', () => expect(formatFileSize(1048576)).toBe('1.0 MB'));
  it('gigabytes', () => expect(formatFileSize(1073741824)).toBe('1.0 GB'));
});

describe('fileExtension', () => {
  it('returns lowercase extension', () => expect(fileExtension('photo.JPG')).toBe('jpg'));
  it('returns empty for no extension', () => expect(fileExtension('filename')).toBe(''));
  it('handles multiple dots', () => expect(fileExtension('archive.tar.gz')).toBe('gz'));
});

describe('isAllowedExtension', () => {
  it('returns true for allowed extension', () => expect(isAllowedExtension('photo.png', ['png', 'jpg'])).toBe(true));
  it('returns false for disallowed extension', () => expect(isAllowedExtension('script.exe', ['png', 'jpg'])).toBe(false));
  it('handles dot prefix in allowed list', () => expect(isAllowedExtension('doc.pdf', ['.pdf'])).toBe(true));
});

describe('isWithinSizeLimit', () => {
  it('returns true when within limit', () => expect(isWithinSizeLimit(500, 1000)).toBe(true));
  it('returns true at exact limit', () => expect(isWithinSizeLimit(1000, 1000)).toBe(true));
  it('returns false when over limit', () => expect(isWithinSizeLimit(1001, 1000)).toBe(false));
});

describe('sanitizeFilename', () => {
  it('replaces unsafe chars', () => expect(sanitizeFilename('my<file>.txt')).toBe('my_file_.txt'));
  it('replaces spaces with underscores', () => expect(sanitizeFilename('my file.txt')).toBe('my_file.txt'));
  it('collapses multiple underscores', () => expect(sanitizeFilename('a__b.txt')).toBe('a_b.txt'));
});
