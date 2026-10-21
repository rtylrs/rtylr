import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const currentDir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(currentDir, 'label.tsx'), 'utf8');

describe('dash label spacing', () => {
  it('includes bottom spacing so labels do not crowd inputs', () => {
    expect(source).toContain('mb-1.5');
  });
});
