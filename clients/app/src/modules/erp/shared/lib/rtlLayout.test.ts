import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');

function source(path: string): string {
  return readFileSync(join(sourceRoot, path), 'utf8');
}

describe('ERP RTL layout', () => {
  it('uses logical geometry for the filter drawer', () => {
    const drawer = source('shared/components/FilterDrawer.tsx');

    expect(drawer).toContain('insetInlineEnd: 0');
    expect(drawer).toContain('borderInlineStart');
    expect(drawer).toContain('const drawerOffset = isRTL ? -320 : 320');
    expect(drawer).not.toContain('borderLeft:');
    expect(drawer).not.toContain('right-0');
  });

  it('mirrors directional icons when the document is RTL', () => {
    const stylesheet = source('erp.css');
    const breadcrumbs = source('shared/components/Breadcrumbs.tsx');
    const pagination = source('shared/components/PaginationControls.tsx');

    expect(stylesheet).toContain("[dir='rtl'] .erp-directional-icon");
    expect(breadcrumbs).toContain('className="erp-directional-icon"');
    expect(pagination).toContain('className="erp-directional-icon"');
  });
});
