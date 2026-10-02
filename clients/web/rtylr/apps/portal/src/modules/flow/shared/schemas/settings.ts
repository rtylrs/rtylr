import { z } from 'zod';

/**
 * Settings section ids, so `/settings/flow?section=time` is linkable.
 *
 * No `z.coerce`: the router JSON-parses query values already, and coerce's input
 * type is `unknown`, which would let `<Link search={{ section: 42 }}>` compile.
 * `.catch(undefined)` because validateSearch runs before the route renders, so a
 * schema that throws replaces the settings screen with the router error boundary
 * on any stale or hand-edited URL. See docs/routing-search-params.md.
 */
const SETTINGS_SECTION_IDS = [
  'notifications',
  'time',
  'appearance',
  'labels',
  'life',
  'shortcuts',
  'sidebar',
  'bin',
] as const;

export type SettingsSectionId = (typeof SETTINGS_SECTION_IDS)[number];

export const settingsSearchSchema = z.object({
  section: z.enum(SETTINGS_SECTION_IDS).optional().catch(undefined),
});
