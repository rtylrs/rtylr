import { describe, it, expect } from 'vitest';
import { content, type DocContent } from '../index';

describe('content registry', () => {
  it('exports a non-empty content map', () => {
    expect(Object.keys(content).length).toBeGreaterThan(0);
  });

  it('every entry has a non-empty title', () => {
    for (const [slug, doc] of Object.entries(content)) {
      expect(doc.title, `slug "${slug}" is missing a title`).toBeTruthy();
      expect(doc.title.trim().length, `slug "${slug}" title is blank`).toBeGreaterThan(0);
    }
  });

  it('every entry has a non-empty body', () => {
    for (const [slug, doc] of Object.entries(content)) {
      expect(doc.body, `slug "${slug}" is missing a body`).toBeTruthy();
      expect(doc.body.trim().length, `slug "${slug}" body is blank`).toBeGreaterThan(0);
    }
  });

  it('includes required top-level documents', () => {
    const requiredSlugs = ['introduction', 'quick-start', 'faq'];
    for (const slug of requiredSlugs) {
      expect(content[slug], `expected content["${slug}"] to exist`).toBeDefined();
    }
  });

  it('includes all core app docs', () => {
    const appSlugs = ['pos', 'erp', 'hr', 'crm', 'finance', 'flow', 'store', 'book', 'support'];
    for (const slug of appSlugs) {
      expect(content[slug], `expected content["${slug}"] to exist`).toBeDefined();
    }
  });

  it('body strings contain HTML tags (docs are HTML-formatted)', () => {
    const htmlPattern = /<[a-z][\s\S]*?>/i;
    for (const [slug, doc] of Object.entries(content)) {
      expect(htmlPattern.test(doc.body), `slug "${slug}" body does not contain HTML`).toBe(true);
    }
  });
});

describe('DocContent type', () => {
  it('satisfies the DocContent shape', () => {
    const sample: DocContent = { title: 'Test', body: '<p>Hello</p>' };
    expect(sample.title).toBe('Test');
    expect(sample.body).toBe('<p>Hello</p>');
  });
});
