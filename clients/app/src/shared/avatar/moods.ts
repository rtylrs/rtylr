/**
 * RTYLR Moods: twelve reaction faces drawn for RTYLR, owned by Voxire.
 *
 * They carry the mood of well-known internet reactions without copying any
 * meme's artwork, character or photo of a real person. Each face is a plain
 * SVG in a 100 by 100 box: one ground colour from the avatar palette, one skin
 * tone, one ink, a 3px round line. Picked like an illustrated style, with the
 * mood id as the seed.
 */
export const MOODS_STYLE_ID = 'rtylr-moods';

const INK = '#2B2118';
const SKIN = '#F6D7A7';
const SKIN_SHADE = '#E8B98A';

export interface Mood {
  id: string;
  /** English name; the picker shows it as the option's label. */
  label: string;
  ground: string;
  body: string;
}

const face = (cy = 54, r = 31): string => `<circle cx="50" cy="${cy}" r="${r}" fill="${SKIN}"/>`;
const dot = (x: number, y: number, r = 3.2): string =>
  `<circle cx="${x}" cy="${y}" r="${r}" fill="${INK}" stroke="none"/>`;

export const MOODS: readonly Mood[] = [
  {
    id: 'stone-face',
    label: 'Stone face',
    ground: '#475467',
    body: `<path d="M31 30 Q50 18 69 30 L72 86 Q50 96 28 86 Z" fill="#C9C2B6"/>
      <path d="M33 46 L46 46 M54 46 L67 46" stroke-width="4"/>
      <path d="M50 49 L47 66 L54 66" />
      <path d="M41 76 L59 76" />
      <path d="M36 36 L44 34 M56 34 L64 36" stroke="#9E978B"/>`,
  },
  {
    id: 'this-is-fine',
    label: 'This is fine',
    ground: '#C4320A',
    body: `<path d="M8 96 Q12 74 20 84 Q22 66 30 82 L30 100 Z M92 96 Q88 72 80 84 Q78 66 70 82 L70 100 Z" fill="#FAC775" stroke="#EF9F27" stroke-width="2"/>
      ${face(50, 28)}
      <path d="M38 46 Q42 42 46 46 M54 46 Q58 42 62 46" fill="none"/>
      <path d="M40 58 Q50 66 60 58" fill="none"/>
      <rect x="62" y="62" width="14" height="16" rx="3" fill="#FFFFFF"/>
      <path d="M76 66 Q82 70 76 74" fill="none" stroke-width="2.5"/>`,
  },
  {
    id: 'side-eye',
    label: 'Side-eye',
    ground: '#0E7090',
    body: `${face()}
      <path d="M33 47 Q40 43 47 47 M53 47 Q60 43 67 47" fill="none"/>
      <path d="M33 49 L47 49 M53 49 L67 49" stroke-width="2"/>
      ${dot(44, 52)}${dot(64, 52)}
      <path d="M40 68 L60 66" />`,
  },
  {
    id: 'galaxy-brain',
    label: 'Galaxy brain',
    ground: '#4338CA',
    body: `${face(60, 28)}
      <path d="M28 40 Q30 18 50 18 Q70 18 72 40 Q60 34 50 38 Q40 34 28 40 Z" fill="#AFA9EC" stroke="#CECBF6" stroke-width="2"/>
      ${dot(38, 25, 1.8).replace(INK, '#FFFFFF')}${dot(60, 22, 1.4).replace(INK, '#FFFFFF')}${dot(66, 32, 1.6).replace(INK, '#FFFFFF')}
      ${dot(41, 58)}${dot(59, 58)}
      <path d="M42 70 Q50 75 58 70" fill="none"/>`,
  },
  {
    id: 'stonks',
    label: 'Stonks',
    ground: '#027A48',
    body: `${face(58, 28)}
      ${dot(41, 55)}${dot(57, 55)}
      <path d="M40 68 Q49 74 58 66" fill="none"/>
      <path d="M58 40 L68 30 L74 36 L88 16 M88 16 L78 17 M88 16 L87 26" stroke="#C0DD97" stroke-width="5" fill="none"/>`,
  },
  {
    id: 'mind-blown',
    label: 'Mind blown',
    ground: '#B42318',
    body: `<path d="M50 30 L44 8 M38 32 L24 14 M62 32 L76 14 M50 30 L57 6 M34 36 L16 30 M66 36 L84 30" stroke="#FAC775" stroke-width="5"/>
      ${face(60, 28)}
      <circle cx="41" cy="56" r="5.5" fill="#FFFFFF"/><circle cx="59" cy="56" r="5.5" fill="#FFFFFF"/>
      ${dot(41, 56, 2)}${dot(59, 56, 2)}
      <ellipse cx="50" cy="72" rx="5" ry="6" fill="${INK}"/>`,
  },
  {
    id: 'facepalm',
    label: 'Facepalm',
    ground: '#A15C07',
    body: `${face()}
      <path d="M40 70 L60 70" />
      <path d="M26 40 Q48 28 70 42 L72 56 Q50 48 26 52 Z" fill="${SKIN_SHADE}"/>
      <path d="M36 42 L37 50 M46 38 L47 48 M56 39 L57 49" stroke="#C99367" stroke-width="2"/>`,
  },
  {
    id: 'deal-with-it',
    label: 'Deal with it',
    ground: '#1D4ED8',
    body: `${face()}
      <rect x="26" y="44" width="21" height="11" rx="2" fill="${INK}"/>
      <rect x="53" y="44" width="21" height="11" rx="2" fill="${INK}"/>
      <path d="M47 48 L53 48" />
      <rect x="30" y="46" width="5" height="3" fill="#FFFFFF" stroke="none"/>
      <rect x="57" y="46" width="5" height="3" fill="#FFFFFF" stroke="none"/>
      <path d="M40 70 Q52 76 62 66" fill="none"/>`,
  },
  {
    id: 'shrug',
    label: 'Shrug',
    ground: '#A21CAF',
    body: `${face(50, 26)}
      <path d="M36 40 Q41 36 46 40 M54 40 Q59 36 64 40" fill="none"/>
      ${dot(41, 47)}${dot(59, 47)}
      <path d="M43 60 Q50 63 57 60" fill="none"/>
      <path d="M24 66 L12 54 L8 42 M76 66 L88 54 L92 42" fill="none" stroke="${SKIN}" stroke-width="6"/>
      <path d="M4 42 L12 40 M88 40 L96 42" stroke="${SKIN}" stroke-width="5"/>`,
  },
  {
    id: 'smug',
    label: 'Smug',
    ground: '#BE185D',
    body: `${face()}
      <path d="M33 42 L45 45 M55 44 L67 38" />
      <path d="M35 51 L45 52 M55 52 L65 51" stroke-width="4"/>
      <path d="M38 68 Q54 72 64 62" fill="none"/>`,
  },
  {
    id: 'thinking',
    label: 'Thinking',
    ground: '#6D28D9',
    body: `${face()}
      <path d="M33 44 L45 44 M55 41 Q61 35 67 39" fill="none"/>
      ${dot(41, 50)}${dot(61, 48)}
      <path d="M42 68 L56 66" />
      <path d="M32 92 Q34 76 46 78 L52 84 Q46 86 44 92" fill="${SKIN_SHADE}"/>`,
  },
  {
    id: 'nervous-sweat',
    label: 'Nervous sweat',
    ground: '#4D7C0F',
    body: `${face()}
      <circle cx="40" cy="50" r="6" fill="#FFFFFF"/><circle cx="60" cy="50" r="6" fill="#FFFFFF"/>
      ${dot(40, 50, 2.2)}${dot(60, 50, 2.2)}
      <path d="M38 68 L42 65 L46 68 L50 65 L54 68 L58 65 L62 68" fill="none" stroke-width="2.5"/>
      <path d="M76 30 Q83 40 76 46 Q69 40 76 30 Z" fill="#B5D4F4" stroke="#85B7EB" stroke-width="2"/>`,
  },
];

export const isMood = (id: string | undefined): boolean => !!id && MOODS.some((m) => m.id === id);

/** One mood as a standalone SVG document. */
export function moodSvg(id: string): string | undefined {
  const mood = MOODS.find((m) => m.id === id);
  if (!mood) return undefined;
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100" fill="${mood.ground}"/><g fill="none" stroke="${INK}" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">${mood.body}</g></svg>`;
}

/** One mood as an SVG data URI, for an `<img>`. */
export function moodDataUri(id: string): string | undefined {
  const svg = moodSvg(id);
  return svg ? `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}` : undefined;
}
