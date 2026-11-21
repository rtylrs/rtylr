import tags from '../../content/blog/tags.json';

import type { BlogLocale } from './types';

type TagLabels = Record<string, Record<BlogLocale, string>>;

const TAGS = tags as TagLabels;

/** A tag's display name. Tags are slugs in post files and URLs. */
export const tagLabel = (slug: string, locale: BlogLocale): string =>
  TAGS[slug]?.[locale] ?? TAGS[slug]?.en ?? slug;

export const knownTags = (): string[] => Object.keys(TAGS);
