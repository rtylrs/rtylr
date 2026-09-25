import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

import { rawNavItems, rawMenuSubItems, routeKeys } from '../layouts/DefaultLayout';

/**
 * Keys that are built rather than written, and are therefore invisible to
 * `i18n-key-resolution.test.ts`.
 *
 * That guard collects literal `t('a.b')` keys. It cannot see `t(`nav.${key}`)`,
 * and dash has ten such call sites. This is not a per-app copy of it: it covers
 * the case that one structurally cannot.
 *
 * It is written because that blind spot already cost something. `nav.createItem`
 * and `nav.createBanner` existed in no locale, and both screens rendered English
 * to every reader, because the call site was `t(`nav.${key}`)` with an English
 * fallback beside it (rtylr#1266).
 */
const LOCALES = path.resolve(__dirname);
const LANGS = ['en', 'ar', 'fr'] as const;

function load(lang: string): Record<string, unknown> {
  const dir = path.join(LOCALES, lang);
  const tree: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    tree[file.replace('.json', '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return tree;
}

function resolves(tree: Record<string, unknown>, dotted: string): boolean {
  let node: unknown = tree;
  for (const part of dotted.split('.')) {
    if (typeof node !== 'object' || node === null) return false;
    node = (node as Record<string, unknown>)[part];
  }
  return typeof node === 'string';
}

/** Each family: the keys it can produce, derived from the code that produces them. */
const FAMILIES: Record<string, string[]> = {
  // DefaultLayout and Breadcrumbs: t(`common.nav.${key}`)
  nav: [
    ...rawNavItems.map((i) => i.key),
    ...rawMenuSubItems.map((i) => i.key),
    ...Object.values(routeKeys),
  ]
    // 'chat' short-circuits to the Caret header before it reaches t().
    .filter((k) => k !== 'chat')
    .map((k) => `common.nav.${k}`),

  // InviteUser and MemberDetail: t(`common.roleLabels.${roleId}`) and its
  // description twin. menu_manager is added when the menu add-on is enabled.
  roleLabels: ['admin', 'manager', 'member', 'viewer', 'menu_manager'].map(
    (r) => `common.roleLabels.${r}`
  ),
  roleDescriptions: ['admin', 'manager', 'member', 'viewer', 'menu_manager'].map(
    (r) => `common.roleDescriptions.${r}`
  ),

  // Design: t(`design.pages.${key}`)
  designPages: ['about', 'gallery', 'reviews', 'offers'].map((p) => `design.pages.${p}`),
};

describe.each(LANGS)('dash dynamic i18n keys resolve in %s', (lang) => {
  const tree = load(lang);

  it.each(Object.entries(FAMILIES))('%s', (_family, keys) => {
    expect(keys.length).toBeGreaterThan(0);
    expect(keys.filter((k) => !resolves(tree, k))).toEqual([]);
  });
});
