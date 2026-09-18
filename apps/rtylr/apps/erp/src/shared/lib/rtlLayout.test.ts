import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');

function source(path: string): string {
  return readFileSync(join(sourceRoot, path), 'utf8');
}

describe('ERP RTL layout', () => {
  it('uses logical geometry for the shell and filter drawer', () => {
    const layout = source('layouts/DefaultLayout.tsx');
    const drawer = source('shared/components/FilterDrawer.tsx');

    expect(layout).toContain('insetInline: 0');
    expect(layout).toContain('insetInlineStart: sidebarOpen ? 0 : undefined');
    expect(layout).toContain('borderInlineEnd');
    expect(layout).toContain("textAlign: 'start'");
    expect(layout).not.toContain('borderRight:');
    expect(layout).not.toContain('text-left');

    expect(drawer).toContain('insetInlineEnd: 0');
    expect(drawer).toContain('borderInlineStart');
    expect(drawer).toContain('const drawerOffset = isRTL ? -320 : 320');
    expect(drawer).not.toContain('borderLeft:');
    expect(drawer).not.toContain('right-0');
  });

  it('mirrors directional icons when the document is RTL', () => {
    const stylesheet = source('globals.css');
    const breadcrumbs = source('shared/components/Breadcrumbs.tsx');
    const pagination = source('shared/components/PaginationControls.tsx');
    const layout = source('layouts/DefaultLayout.tsx');

    expect(stylesheet).toContain("[dir='rtl'] .erp-directional-icon");
    expect(breadcrumbs).toContain('className="erp-directional-icon"');
    expect(pagination).toContain('className="erp-directional-icon"');
    expect(layout).toContain('className="erp-directional-icon shrink-0"');
  });
});
