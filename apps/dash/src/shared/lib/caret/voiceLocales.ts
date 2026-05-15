/** Supported speech-recognition locales surfaced in the settings UI. */
export interface VoiceLocale {
  code: string;
  label: string;
}

export const VOICE_LOCALES: VoiceLocale[] = [
  { code: 'en-US', label: 'English (US)' },
  { code: 'en-GB', label: 'English (UK)' },
  { code: 'fr-FR', label: 'French' },
  { code: 'es-ES', label: 'Spanish' },
  { code: 'de-DE', label: 'German' },
  { code: 'it-IT', label: 'Italian' },
  { code: 'pt-BR', label: 'Portuguese (Brazil)' },
  { code: 'ar-LB', label: 'Arabic (Lebanon)' },
  { code: 'ar-SA', label: 'Arabic (Saudi)' },
  { code: 'ja-JP', label: 'Japanese' },
];

export const DEFAULT_VOICE_LOCALE = 'en-US';
