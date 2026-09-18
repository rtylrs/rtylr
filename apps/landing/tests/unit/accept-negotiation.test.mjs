import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { parseAccept, negotiate } from '../../apps/landing/cloudflare/_worker.js';

const PRODUCES = ['text/html', 'text/markdown'];

/**
 * Test vectors lifted verbatim from acceptmarkdown.com/guides/accept-parsing
 * ("Quick sanity checks your parser should pass"), plus RFC 9110 edge cases.
 */
describe('parseAccept', () => {
  it('returns [] for absent or empty headers', () => {
    assert.deepEqual(parseAccept(null), []);
    assert.deepEqual(parseAccept(undefined), []);
    assert.deepEqual(parseAccept(''), []);
    assert.deepEqual(parseAccept('   '), []);
  });

  it('lowercases types and defaults q to 1', () => {
    const entries = parseAccept('TEXT/Markdown ; Q=0.5');
    assert.deepEqual(entries, [{ type: 'text/markdown', q: 0.5, specificity: 2 }]);
  });

  it('clamps out-of-range q values into [0,1]', () => {
    assert.equal(parseAccept('text/html;q=7')[0].q, 1);
    assert.equal(parseAccept('text/html;q=-2')[0].q, 0);
  });

  it('ignores malformed entries without losing the rest', () => {
    const entries = parseAccept(';q=1, text/markdown');
    assert.equal(entries.length, 1);
    assert.equal(entries[0].type, 'text/markdown');
  });

  it('ignores junk that is not a media range', () => {
    assert.deepEqual(parseAccept('garbage'), []);
    assert.deepEqual(parseAccept('<>'), []);
  });
});

describe('negotiate (acceptmarkdown.com test vectors)', () => {
  const cases = [
    // [Accept header, expected choice]
    ['text/markdown', 'text/markdown'],
    ['text/markdown, text/html;q=0.8', 'text/markdown'],
    ['text/html', 'text/html'],
    ['text/markdown;q=0, text/html', 'text/html'],
    ['(no Accept)', 'text/html'], // missing header = no constraint
    ['*/*', 'text/html'], // anything is fine -> server default
    ['', 'text/html'],
  ];

  for (const [accept, expected] of cases) {
    const header = accept === '(no Accept)' ? null : accept;
    it(`Accept: ${accept} -> ${expected.replace(/^text\//, '')}`, () => {
      assert.equal(negotiate(header, PRODUCES), expected);
    });
  }

  it('rejects everything when the client disallows all producible types', () => {
    assert.equal(negotiate('application/pdf', PRODUCES), null);
    assert.equal(negotiate('text/*;q=0', PRODUCES), null);
    assert.equal(negotiate('*/*;q=0', PRODUCES), null);
    assert.equal(
      negotiate('text/markdown;q=0, text/html;q=0', PRODUCES),
      null,
      'q=0 on every candidate must yield null (406)'
    );
  });

  it('a specific q=0 overrides a wildcard allow (RFC 9110 specificity rule)', () => {
    assert.equal(negotiate('text/html;q=0, */*;q=1', PRODUCES), 'text/markdown');
    assert.equal(negotiate('text/markdown;q=0, */*;q=1', PRODUCES), 'text/html');
  });

  it('breaks q ties using client ordering', () => {
    assert.equal(negotiate('text/markdown, text/html', PRODUCES), 'text/markdown');
    assert.equal(negotiate('text/html, text/markdown', PRODUCES), 'text/html');
  });

  it('ranks strictly by q-value', () => {
    assert.equal(negotiate('text/markdown;q=0.4, text/html;q=0.6', PRODUCES), 'text/html');
    assert.equal(negotiate('text/markdown;q=0.9, text/html;q=0.1', PRODUCES), 'text/markdown');
  });
});
