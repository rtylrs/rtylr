export interface LanguageDefinition {
  /** BCP-47 language tag used as i18next lng key */
  code: string;
  /** Name in English */
  name: string;
  /** Name written in the language itself */
  nativeName: string;
  /** Right-to-left script */
  rtl: boolean;
}

export const SUPPORTED_LANGUAGES: LanguageDefinition[] = [
  { code: 'en', name: 'English', nativeName: 'English', rtl: false },
  { code: 'fr', name: 'French', nativeName: 'Français', rtl: false },
  { code: 'ar', name: 'Arabic', nativeName: 'العربية', rtl: true },
  { code: 'es', name: 'Spanish', nativeName: 'Español', rtl: false },
  { code: 'pt', name: 'Portuguese', nativeName: 'Português', rtl: false },
  { code: 'de', name: 'German', nativeName: 'Deutsch', rtl: false },
  { code: 'it', name: 'Italian', nativeName: 'Italiano', rtl: false },
  { code: 'nl', name: 'Dutch', nativeName: 'Nederlands', rtl: false },
  { code: 'tr', name: 'Turkish', nativeName: 'Türkçe', rtl: false },
  { code: 'pl', name: 'Polish', nativeName: 'Polski', rtl: false },
  { code: 'ru', name: 'Russian', nativeName: 'Русский', rtl: false },
  { code: 'uk', name: 'Ukrainian', nativeName: 'Українська', rtl: false },
  { code: 'zh-CN', name: 'Chinese (Simplified)', nativeName: '中文（简体）', rtl: false },
  { code: 'zh-TW', name: 'Chinese (Traditional)', nativeName: '中文（繁體）', rtl: false },
  { code: 'ja', name: 'Japanese', nativeName: '日本語', rtl: false },
  { code: 'ko', name: 'Korean', nativeName: '한국어', rtl: false },
  { code: 'hi', name: 'Hindi', nativeName: 'हिन्दी', rtl: false },
  { code: 'bn', name: 'Bengali', nativeName: 'বাংলা', rtl: false },
  { code: 'ur', name: 'Urdu', nativeName: 'اردو', rtl: true },
  { code: 'he', name: 'Hebrew', nativeName: 'עברית', rtl: true },
  { code: 'fa', name: 'Persian', nativeName: 'فارسی', rtl: true },
  { code: 'id', name: 'Indonesian', nativeName: 'Bahasa Indonesia', rtl: false },
  { code: 'ms', name: 'Malay', nativeName: 'Bahasa Melayu', rtl: false },
  { code: 'vi', name: 'Vietnamese', nativeName: 'Tiếng Việt', rtl: false },
  { code: 'th', name: 'Thai', nativeName: 'ภาษาไทย', rtl: false },
  { code: 'fil', name: 'Filipino', nativeName: 'Filipino', rtl: false },
  { code: 'ro', name: 'Romanian', nativeName: 'Română', rtl: false },
  { code: 'cs', name: 'Czech', nativeName: 'Čeština', rtl: false },
  { code: 'hu', name: 'Hungarian', nativeName: 'Magyar', rtl: false },
  { code: 'sv', name: 'Swedish', nativeName: 'Svenska', rtl: false },
  { code: 'no', name: 'Norwegian', nativeName: 'Norsk', rtl: false },
  { code: 'da', name: 'Danish', nativeName: 'Dansk', rtl: false },
  { code: 'fi', name: 'Finnish', nativeName: 'Suomi', rtl: false },
  { code: 'el', name: 'Greek', nativeName: 'Ελληνικά', rtl: false },
  { code: 'sw', name: 'Swahili', nativeName: 'Kiswahili', rtl: false },
];

export const RTL_LANGUAGES = new Set(SUPPORTED_LANGUAGES.filter((l) => l.rtl).map((l) => l.code));

/**
 * Whether a BCP-47 tag is written right-to-left.
 *
 * Matching is on the primary subtag, so a regional tag inherits its base
 * language's direction. An exact match alone answered `false` for `ar-EG` and
 * `ar-SA`, which lays an Arabic page out left-to-right; the language cookie is
 * shared across apps on one domain, so a regional tag from any of them reaches
 * here. Script subtags are not resolved, so `ar-Latn` is reported as RTL.
 */
export function isRTL(lang: string): boolean {
  const tag = lang.trim().toLowerCase().replace(/_/g, '-');
  if (RTL_LANGUAGES.has(tag)) return true;
  return RTL_LANGUAGES.has(tag.split('-')[0]!);
}

export function getLanguage(code: string): LanguageDefinition | undefined {
  return SUPPORTED_LANGUAGES.find((l) => l.code === code);
}
