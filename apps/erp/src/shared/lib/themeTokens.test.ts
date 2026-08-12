import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');
const screenRoots = ['features', 'layouts', 'shared/components'];
const fixedOutputFiles = new Set([
  'shared/components/BarcodeSvg.tsx',
  'shared/components/PrintPOButton.tsx',
]);
const rawColorPattern = /#[0-9a-f]{3,8}\b|rgba?\s*\(/gi;

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
    const stylesheet = readFileSync(join(sourceRoot, 'globals.css'), 'utf8');

    for (const token of [
      '--erp-accent',
      '--erp-accent-foreground',
      '--erp-success',
      '--erp-danger',
      '--erp-info',
      '--erp-purple',
      '--hub-overlay',
    ]) {
      expect(stylesheet).toContain(`${token}:`);
    }
  });

  it('keeps raw colors out of screen components', () => {
    const violations = screenRoots
      .flatMap((root) => sourceFiles(join(sourceRoot, root)))
      .filter((file) => !fixedOutputFiles.has(file.slice(`${sourceRoot}/`.length)))
      .flatMap((file) => {
        const source = readFileSync(file, 'utf8');
        return source.match(rawColorPattern)
          ? [`${file}: ${source.match(rawColorPattern)?.join(', ')}`]
          : [];
      });

    expect(violations).toEqual([]);
  });
});
