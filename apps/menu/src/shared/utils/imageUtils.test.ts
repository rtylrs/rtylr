import { describe, it, expect } from 'vitest';
import {
  isImageUrl,
  placeholderImage,
  menuItemAlt,
  buildSrcSet,
  cssBackgroundImage,
  clampDimension,
  pxToRem,
  aspectRatioCss,
} from './imageUtils';

describe('isImageUrl', () => {
  it('returns true for common image extensions', () => {
    expect(isImageUrl('https://example.com/photo.jpg')).toBe(true);
    expect(isImageUrl('https://example.com/photo.png')).toBe(true);
    expect(isImageUrl('https://example.com/photo.webp')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(isImageUrl('https://example.com/photo.JPG')).toBe(true);
  });

  it('handles query strings', () => {
    expect(isImageUrl('https://example.com/photo.jpg?w=800')).toBe(true);
  });

  it('returns false for non-image URLs', () => {
    expect(isImageUrl('https://example.com/page.html')).toBe(false);
  });

  it('returns false for empty string', () => {
    expect(isImageUrl('')).toBe(false);
  });
});

describe('placeholderImage', () => {
  it('returns a URL with default dimensions', () => {
    expect(placeholderImage()).toContain('400x300');
  });

  it('accepts custom dimensions', () => {
    expect(placeholderImage(800, 600)).toContain('800x600');
  });
});

describe('menuItemAlt', () => {
  it('returns alt text with item name', () => {
    expect(menuItemAlt('Spicy Burger')).toBe('Photo of Spicy Burger');
  });
});

describe('buildSrcSet', () => {
  it('builds srcset string from widths', () => {
    const result = buildSrcSet('https://img.example.com/photo.jpg', [400, 800]);
    expect(result).toBe(
      'https://img.example.com/photo.jpg?w=400 400w, https://img.example.com/photo.jpg?w=800 800w'
    );
  });

  it('returns empty string for empty widths', () => {
    expect(buildSrcSet('https://img.example.com/photo.jpg', [])).toBe('');
  });
});

describe('cssBackgroundImage', () => {
  it('wraps URL in CSS function', () => {
    expect(cssBackgroundImage('https://example.com/img.jpg')).toBe(
      'url("https://example.com/img.jpg")'
    );
  });
});

describe('clampDimension', () => {
  it('clamps below min', () => {
    expect(clampDimension(10, 50, 200)).toBe(50);
  });

  it('clamps above max', () => {
    expect(clampDimension(500, 50, 200)).toBe(200);
  });

  it('returns value within range unchanged', () => {
    expect(clampDimension(100, 50, 200)).toBe(100);
  });
});

describe('pxToRem', () => {
  it('converts pixels to rem', () => {
    expect(pxToRem(16)).toBe('1rem');
    expect(pxToRem(32)).toBe('2rem');
  });

  it('accepts a custom base', () => {
    expect(pxToRem(20, 10)).toBe('2rem');
  });
});

describe('aspectRatioCss', () => {
  it('returns aspect ratio string', () => {
    expect(aspectRatioCss(16, 9)).toBe('16 / 9');
  });

  it('handles square ratio', () => {
    expect(aspectRatioCss(1, 1)).toBe('1 / 1');
  });
});
