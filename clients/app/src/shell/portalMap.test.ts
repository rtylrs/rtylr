import fs from 'node:fs';
import path from 'node:path';
import type { TFunction } from 'i18next';
import { describe, expect, it } from 'vitest';
import { createEffectiveAccess } from '../../test/effectiveAccess';
import type { StoredUser } from '../shared/stores/userStore';
import { router } from '../router';
import { isPortalRoute } from '../features/caret/portalPlaces';
import { rawMenuSubItems, rawNavItems } from './navigation';
import { arrangeItems } from './portalLayout';
import { RAIL_BOTTOM, railItems } from './railItems';

// The server's map of the Portal: Caret finds pages and arranges the
// navigation from it, so it has to describe this Portal exactly.
const catalog = path.resolve(
  import.meta.dirname,
  '../../../../services/onion/internal/modules/portal/internal/catalog'
);
const read = <T>(file: string): T =>
  JSON.parse(fs.readFileSync(path.join(catalog, file), 'utf8')) as T;
const navigation = read<{
  rail: { items: { id: string; bottom?: boolean }[] };
  workspace: { items: { id: string }[] };
}>('navigation.json');
const pages = read<{ path: string }[]>('pages.json');

const key = ((value: string) => value) as unknown as TFunction;
const owner = {
  id: 1,
  effectiveAccess: createEffectiveAccess(undefined, {
    appIds: ['insights', 'flow', 'crm', 'erp', 'finance', 'hr', 'pos'],
    isOwner: true,
  }),
} as unknown as StoredUser;

describe('Portal map', () => {
  it('names the app bar items in the order the Portal shows them', () => {
    expect(railItems(owner, key).map((item) => item.id)).toEqual(
      navigation.rail.items.map((item) => item.id)
    );
    expect(RAIL_BOTTOM).toEqual(
      navigation.rail.items.filter((item) => item.bottom).map((item) => item.id)
    );
  });

  it('names the Home sidebar items in the order the Portal shows them', () => {
    const shown = [...rawNavItems, ...rawMenuSubItems]
      .sort((a, b) => Number(a.key === 'extensions') - Number(b.key === 'extensions'))
      .map((item) => item.key);
    expect(shown).toEqual(navigation.workspace.items.map((item) => item.id));
  });

  it('only lists pages the Portal has', () => {
    expect(pages.length).toBeGreaterThan(0);
    const missing = pages.filter((page) => !isPortalRoute(router, page.path));
    expect(missing.map((page) => page.path)).toEqual([]);
    expect(isPortalRoute(router, '/not-a-portal-page')).toBe(false);
  });
});

describe('arrangeItems', () => {
  const items = ['home', 'crm', 'hr', 'caret'].map((id) => ({ id }));

  it('puts Caret at the top when the person moved it there', () => {
    const { top, bottom } = arrangeItems(
      items,
      { top: ['caret', 'home', 'crm', 'hr'], bottom: [] },
      RAIL_BOTTOM
    );
    expect(top.map((item) => item.id)).toEqual(['caret', 'home', 'crm', 'hr']);
    expect(bottom).toEqual([]);
  });

  it('drops hidden and unknown items and keeps new ones in their default group', () => {
    const { top, bottom } = arrangeItems(
      items,
      { top: ['gone', 'hr', 'home'], bottom: [], hidden: ['hr'] },
      RAIL_BOTTOM
    );
    expect(top.map((item) => item.id)).toEqual(['home', 'crm']);
    expect(bottom.map((item) => item.id)).toEqual(['caret']);
  });

  it('shows the default order before the layout loads', () => {
    const { top, bottom } = arrangeItems(items, undefined, RAIL_BOTTOM);
    expect(top.map((item) => item.id)).toEqual(['home', 'crm', 'hr']);
    expect(bottom.map((item) => item.id)).toEqual(['caret']);
  });
});
