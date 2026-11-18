/**
 * A user's chosen avatar, as the hub user API carries it (`UserAvatar`).
 * Absent means the default: initials on the colour their id picks.
 */
export interface ChosenAvatar {
  kind: 'photo' | 'illustrated' | 'initials' | 'emoji';
  url?: string;
  style?: string;
  seed?: string;
  /** Index into AVATAR_PALETTE (@voxire/ui). */
  color?: number;
  emoji?: string;
}

export type AvatarStyleGroup = 'professional' | 'cool' | 'funky' | 'abstract';

export interface AvatarStyleCredit {
  author: string;
  work: string;
  workUrl: string;
  license: 'CC BY 4.0' | 'Free for personal and commercial use';
  licenseUrl: string;
}

export interface AvatarStyle {
  id: string;
  label: string;
  group: AvatarStyleGroup;
  /** Shown under the style in the picker. CC0 styles need none. */
  credit?: AvatarStyleCredit;
  /** Released by DiceBear in 2026. */
  isNew?: boolean;
}

const CC_BY = 'https://creativecommons.org/licenses/by/4.0/';

/**
 * The twenty styles Abed picked (2026-11-18), plus RTYLR Moods, by mood. Licences from
 * dicebear.com/licenses: CC BY 4.0 styles carry a credit line in the picker,
 * which is all that licence asks; Avataaars is the artist's own free terms;
 * the rest are CC0. Keep this list and onion's `domain.AvatarStyles` in step:
 * the server refuses a style it does not know.
 */
export const AVATAR_STYLES: readonly AvatarStyle[] = [
  { id: 'notionists', label: 'Notionists', group: 'professional' },
  { id: 'lorelei', label: 'Lorelei', group: 'professional' },
  {
    id: 'personas',
    label: 'Personas',
    group: 'professional',
    credit: {
      author: 'Draftbit',
      work: 'Personas by Draftbit',
      workUrl: 'https://personas.draftbit.com/',
      license: 'CC BY 4.0',
      licenseUrl: CC_BY,
    },
  },
  {
    id: 'micah',
    label: 'Micah',
    group: 'professional',
    credit: {
      author: 'Micah Lanier',
      work: 'Avatar Illustration System',
      workUrl: 'https://www.figma.com/community/file/829741575478342595',
      license: 'CC BY 4.0',
      licenseUrl: CC_BY,
    },
  },
  { id: 'cameo', label: 'Cameo', group: 'professional', isNew: true },
  { id: 'shadows', label: 'Shadows', group: 'professional', isNew: true },
  {
    id: 'avataaars',
    label: 'Avataaars',
    group: 'cool',
    credit: {
      author: 'Pablo Stanley',
      work: 'Avataaars',
      workUrl: 'https://avataaars.com/',
      license: 'Free for personal and commercial use',
      licenseUrl: 'https://avataaars.com/',
    },
  },
  {
    id: 'adventurer',
    label: 'Adventurer',
    group: 'cool',
    credit: {
      author: 'Lisa Wischofsky',
      work: 'Adventurer',
      workUrl: 'https://www.figma.com/community/file/1184595184137881796',
      license: 'CC BY 4.0',
      licenseUrl: CC_BY,
    },
  },
  {
    id: 'big-smile',
    label: 'Big Smile',
    group: 'cool',
    credit: {
      author: 'Ashley Seo',
      work: 'Custom Avatar',
      workUrl: 'https://www.figma.com/community/file/881358461963645496',
      license: 'CC BY 4.0',
      licenseUrl: CC_BY,
    },
  },
  { id: 'voxel-art', label: 'Voxel Art', group: 'cool', isNew: true },
  { id: 'cutouts', label: 'Cutouts', group: 'cool', isNew: true },
  { id: 'line-face', label: 'Line Face', group: 'cool', isNew: true },
  { id: 'thumbs', label: 'Thumbs', group: 'funky' },
  {
    id: 'croodles',
    label: 'Croodles',
    group: 'funky',
    credit: {
      author: 'vijay verma',
      work: 'Croodles - Doodle your face',
      workUrl: 'https://www.figma.com/community/file/966199982810283152',
      license: 'CC BY 4.0',
      licenseUrl: CC_BY,
    },
  },
  { id: 'voxel-bot', label: 'Voxel Bot', group: 'funky', isNew: true },
  { id: 'gaze', label: 'Gaze', group: 'funky', isNew: true },
  // Drawn for RTYLR and owned by Voxire: see moods.ts.
  { id: 'rtylr-moods', label: 'RTYLR Moods', group: 'funky', isNew: true },
  { id: 'glass', label: 'Glass', group: 'abstract' },
  { id: 'slice', label: 'Slice', group: 'abstract', isNew: true },
  { id: 'stack', label: 'Stack', group: 'abstract', isNew: true },
  { id: 'patchwork', label: 'Patchwork', group: 'abstract', isNew: true },
];

export const AVATAR_STYLE_GROUPS: readonly AvatarStyleGroup[] = [
  'professional',
  'cool',
  'funky',
  'abstract',
];

export const isKnownStyle = (id: string | undefined): boolean =>
  !!id && AVATAR_STYLES.some((s) => s.id === id);

/**
 * Reaction faces first (the "meme" mood without anyone's artwork), then
 * everyday ones. Each is a single emoji the server accepts.
 */
export const AVATAR_EMOJI: readonly string[] = [
  '🗿',
  '💀',
  '🫠',
  '🤡',
  '😎',
  '🤌',
  '🙃',
  '🥲',
  '😏',
  '🤯',
  '🫡',
  '🧐',
  '🦊',
  '🐼',
  '🐨',
  '🦁',
  '🐯',
  '🐸',
  '🐙',
  '🦄',
  '🐝',
  '🦉',
  '🐳',
  '🐧',
  '🚀',
  '🌵',
  '🍕',
  '☕',
  '🎧',
  '🎨',
  '⚡',
  '🔥',
  '🌈',
  '⭐',
  '🍀',
  '🎯',
];

/** A fresh seed for Shuffle: short, URL-safe, and what the server accepts. */
export function randomAvatarSeed(): string {
  const bytes = new Uint8Array(6);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(36).padStart(2, '0'))
    .join('')
    .slice(0, 10);
}
