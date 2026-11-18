import type { StyleDefinition } from '@dicebear/core';
import { MOODS_STYLE_ID, moodDataUri } from './moods';

/**
 * Style definitions are 7 to 370 KB of JSON each, so each one is its own chunk
 * and only loads when an avatar in that style is on screen.
 */
const LOADERS: Record<string, () => Promise<{ default: unknown }>> = {
  notionists: () => import('@dicebear/styles/notionists.json'),
  lorelei: () => import('@dicebear/styles/lorelei.json'),
  personas: () => import('@dicebear/styles/personas.json'),
  micah: () => import('@dicebear/styles/micah.json'),
  avataaars: () => import('@dicebear/styles/avataaars.json'),
  adventurer: () => import('@dicebear/styles/adventurer.json'),
  'big-smile': () => import('@dicebear/styles/big-smile.json'),
  thumbs: () => import('@dicebear/styles/thumbs.json'),
  croodles: () => import('@dicebear/styles/croodles.json'),
  glass: () => import('@dicebear/styles/glass.json'),
  cameo: () => import('@dicebear/styles/cameo.json'),
  shadows: () => import('@dicebear/styles/shadows.json'),
  cutouts: () => import('@dicebear/styles/cutouts.json'),
  'voxel-art': () => import('@dicebear/styles/voxel-art.json'),
  'line-face': () => import('@dicebear/styles/line-face.json'),
  'voxel-bot': () => import('@dicebear/styles/voxel-bot.json'),
  slice: () => import('@dicebear/styles/slice.json'),
  gaze: () => import('@dicebear/styles/gaze.json'),
  stack: () => import('@dicebear/styles/stack.json'),
  patchwork: () => import('@dicebear/styles/patchwork.json'),
};

const rendered = new Map<string, string>();
const pending = new Map<string, Promise<string>>();

const key = (style: string, seed: string): string => `${style}:${seed}`;

/** A data URI already rendered for this style and seed, if any. */
export function cachedIllustration(style: string, seed: string): string | undefined {
  if (style === MOODS_STYLE_ID) return moodDataUri(seed);
  return rendered.get(key(style, seed));
}

/**
 * Renders an illustrated avatar as an SVG data URI. Shown through `<img>`, so
 * the SVG's internal ids cannot collide across avatars on one page, and its
 * markup never enters the document.
 */
export function renderIllustration(style: string, seed: string): Promise<string> {
  if (style === MOODS_STYLE_ID) {
    const mood = moodDataUri(seed);
    return mood ? Promise.resolve(mood) : Promise.reject(new Error(`Unknown mood: ${seed}`));
  }
  const k = key(style, seed);
  const done = rendered.get(k);
  if (done) return Promise.resolve(done);
  const inFlight = pending.get(k);
  if (inFlight) return inFlight;
  const load = LOADERS[style];
  if (!load) return Promise.reject(new Error(`Unknown avatar style: ${style}`));
  const promise = Promise.all([import('@dicebear/core'), load()])
    .then(([core, definition]) => {
      const avatar = new core.Avatar(new core.Style(definition.default as StyleDefinition), {
        seed,
      });
      const uri = avatar.toDataUri();
      rendered.set(k, uri);
      return uri;
    })
    .finally(() => pending.delete(k));
  pending.set(k, promise);
  return promise;
}
