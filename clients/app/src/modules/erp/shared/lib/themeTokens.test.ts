import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');
const screenRoots = ['features', 'shared/components'];
const fixedOutputFiles = new Set([
  'shared/components/BarcodeSvg.tsx',
  'shared/components/PrintPOButton.tsx',
]);
const rawColorPattern = /#[0-9a-f]{3,8}\b|rgba?\s*\(/gi;

/**
 * A line that is only a comment. A colour that does not render is not a raw
 * colour, and more to the point an issue reference is shaped exactly like a
 * short hex: `#1269` and `#988` both match the pattern above, so before this
 * every mention of an issue number in an erp comment failed this test.
 */
const commentLine = /^\s*(\/\/|\*|\/\*)/;

function withoutComments(source: string): string {
  return source
    .split('\n')
    .filter((line) => !commentLine.test(line))
    .join('\n');
}

function sourceFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const relativePath = entry.name;
    const absolutePath = join(directory, relativePath);
    if (entry.isDirectory()) return sourceFiles(absolutePath);
    return /\.(ts|tsx)$/.test(entry.name) ? [absolutePath] : [];
  });
}

describe('ERP theme tokens', () => {
  it('defines the semantic palette used by ERP screens', () => {
    const stylesheet = readFileSync(join(sourceRoot, 'erp.css'), 'utf8');

    for (const token of [
      '--erp-accent',
      '--erp-accent-foreground',
      '--erp-success',
      '--erp-danger',
      '--erp-info',
      '--erp-purple',
    ]) {
      expect(stylesheet).toContain(`${token}:`);
    }
  });

  it('keeps raw colors out of screen components', () => {
    const violations = screenRoots
      .flatMap((root) => sourceFiles(join(sourceRoot, root)))
      .filter((file) => !fixedOutputFiles.has(file.slice(`${sourceRoot}/`.length)))
      .flatMap((file) => {
        const source = withoutComments(readFileSync(file, 'utf8'));
        return source.match(rawColorPattern)
          ? [`${file}: ${source.match(rawColorPattern)?.join(', ')}`]
          : [];
      });

    expect(violations).toEqual([]);
  });
});
