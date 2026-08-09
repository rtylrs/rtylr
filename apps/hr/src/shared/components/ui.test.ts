import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const source = readFileSync(join(import.meta.dirname, 'ui.tsx'), 'utf8');

describe('hr field spacing', () => {
  it('keeps a bit more space between labels and controls', () => {
    expect(source).toContain('marginBottom: 8');
  });
});
