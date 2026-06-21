/** Color presets users can pick to retheme the Caret accent. */
export interface AccentPreset {
  id: 'teal' | 'violet' | 'sky' | 'emerald' | 'amber' | 'rose';
  label: string;
  accent: string;
  accentSoft: string;
}

export const ACCENT_PRESETS: AccentPreset[] = [
  { id: 'teal', label: 'Teal', accent: '#64c5b9', accentSoft: 'rgba(100,197,185,0.14)' },
  { id: 'violet', label: 'Violet', accent: '#8b5cf6', accentSoft: 'rgba(139,92,246,0.14)' },
  { id: 'sky', label: 'Sky', accent: '#0ea5e9', accentSoft: 'rgba(14,165,233,0.14)' },
  { id: 'emerald', label: 'Emerald', accent: '#10b981', accentSoft: 'rgba(16,185,129,0.14)' },
  { id: 'amber', label: 'Amber', accent: '#f59e0b', accentSoft: 'rgba(245,158,11,0.14)' },
  { id: 'rose', label: 'Rose', accent: '#f43f5e', accentSoft: 'rgba(244,63,94,0.14)' },
];

export const DEFAULT_ACCENT: AccentPreset['id'] = 'teal';
