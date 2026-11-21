/**
 * The words a scanned phone shows, by the key the pager service sends. The
 * phone and the staff board both read this list, so the customer and the
 * cashier always see the same emoji and word. The service keeps the same keys
 * in the same order (services/onion/internal/modules/pager/internal/domain/match_word.go);
 * matchWords.test.ts fails when the two differ.
 *
 * Slang words keep their English spelling in French and are transliterated in
 * Arabic. No word is rude or has a sexual, drug or alcohol meaning, and no two
 * sound alike when said out loud. Every emoji is from Unicode 11 or earlier, so
 * it renders on current iPhones and Android phones.
 */
export interface MatchWord {
  key: string;
  emoji: string;
  en: string;
  fr: string;
  ar: string;
}

const word = (key: string, emoji: string, en: string, fr: string, ar: string): MatchWord => ({
  key,
  emoji,
  en,
  fr,
  ar,
});

const slang = (key: string, emoji: string, en: string, ar: string): MatchWord =>
  word(key, emoji, en, en, ar);

export const MATCH_WORDS: readonly MatchWord[] = [
  slang('aura', '✨', 'Aura', 'أورا'),
  slang('slay', '💅', 'Slay', 'سلاي'),
  slang('yeet', '🚀', 'Yeet', 'ييت'),
  slang('bestie', '💞', 'Bestie', 'بيستي'),
  slang('vibe', '🌈', 'Vibe', 'فايب'),
  slang('rizz', '😎', 'Rizz', 'ريز'),
  slang('delulu', '🦄', 'Delulu', 'ديلولو'),
  slang('goat', '🐐', 'GOAT', 'غوت'),
  slang('sidequest', '🗺️', 'Side quest', 'سايد كويست'),
  slang('drip', '💧', 'Drip', 'دريب'),
  slang('bussin', '🤤', 'Bussin', 'باسن'),
  slang('sus', '🤨', 'Sus', 'ساس'),
  slang('glowup', '🌟', 'Glow up', 'غلو أب'),
  slang('chefskiss', '😘', "Chef's kiss", 'شيفز كيس'),
  slang('nocap', '🧢', 'No cap', 'نو كاب'),
  slang('bet', '🤝', 'Bet', 'بِت'),
  slang('iconic', '👑', 'Iconic', 'أيكونيك'),
  slang('lowkey', '🤫', 'Lowkey', 'لو كي'),
  slang('mood', '🎭', 'Mood', 'مود'),
  slang('hype', '🎉', 'Hype', 'هايب'),
  slang('legend', '🦸', 'Legend', 'ليجند'),
  // The ice cube emoji is Unicode 12, too new for some phones.
  slang('chill', '😌', 'Chill', 'تشيل'),
  slang('tea', '☕', 'Tea', 'تي'),
  slang('cringe', '😬', 'Cringe', 'كرينج'),
  slang('yap', '🗣️', 'Yap', 'ياب'),
  slang('era', '💿', 'Era', 'إيرا'),
  slang('lockedin', '🔒', 'Locked in', 'لوكد إن'),
  slang('npc', '🤖', 'NPC', 'إن بي سي'),
  word('panda', '🐼', 'Panda', 'Panda', 'باندا'),
  word('lion', '🦁', 'Lion', 'Lion', 'أسد'),
  word('tiger', '🐯', 'Tiger', 'Tigre', 'نمر'),
  word('koala', '🐨', 'Koala', 'Koala', 'كوالا'),
  word('rabbit', '🐰', 'Rabbit', 'Lapin', 'أرنب'),
  word('fox', '🦊', 'Fox', 'Renard', 'ثعلب'),
  word('penguin', '🐧', 'Penguin', 'Pingouin', 'بطريق'),
  word('frog', '🐸', 'Frog', 'Grenouille', 'ضفدع'),
  word('turtle', '🐢', 'Turtle', 'Tortue', 'سلحفاة'),
  word('octopus', '🐙', 'Octopus', 'Pieuvre', 'أخطبوط'),
  word('whale', '🐳', 'Whale', 'Baleine', 'حوت'),
  word('dolphin', '🐬', 'Dolphin', 'Dauphin', 'دلفين'),
  word('butterfly', '🦋', 'Butterfly', 'Papillon', 'فراشة'),
  word('horse', '🐴', 'Horse', 'Cheval', 'حصان'),
  word('giraffe', '🦒', 'Giraffe', 'Girafe', 'زرافة'),
  word('elephant', '🐘', 'Elephant', 'Éléphant', 'فيل'),
  word('camel', '🐪', 'Camel', 'Chameau', 'جمل'),
  word('hedgehog', '🦔', 'Hedgehog', 'Hérisson', 'قنفذ'),
  word('crab', '🦀', 'Crab', 'Crabe', 'سلطعون'),
  word('shark', '🦈', 'Shark', 'Requin', 'قرش'),
  word('kangaroo', '🦘', 'Kangaroo', 'Kangourou', 'كنغر'),
  word('wolf', '🐺', 'Wolf', 'Loup', 'ذئب'),
  word('cat', '🐱', 'Cat', 'Chat', 'قطة'),
  word('crocodile', '🐊', 'Crocodile', 'Crocodile', 'تمساح'),
  word('hamster', '🐹', 'Hamster', 'Hamster', 'هامستر'),
  word('dinosaur', '🦖', 'Dinosaur', 'Dinosaure', 'ديناصور'),
  word('banana', '🍌', 'Banana', 'Banane', 'موز'),
  word('grapes', '🍇', 'Grapes', 'Raisin', 'عنب'),
  word('strawberry', '🍓', 'Strawberry', 'Fraise', 'فراولة'),
  word('cherry', '🍒', 'Cherry', 'Cerise', 'كرز'),
  word('lemon', '🍋', 'Lemon', 'Citron', 'ليمون'),
  word('pineapple', '🍍', 'Pineapple', 'Ananas', 'أناناس'),
  word('mango', '🥭', 'Mango', 'Mangue', 'مانجو'),
  word('coconut', '🥥', 'Coconut', 'Noix de coco', 'جوز الهند'),
  word('avocado', '🥑', 'Avocado', 'Avocat', 'أفوكادو'),
  word('carrot', '🥕', 'Carrot', 'Carotte', 'جزرة'),
  word('broccoli', '🥦', 'Broccoli', 'Brocoli', 'بروكلي'),
  word('donut', '🍩', 'Donut', 'Donut', 'دونات'),
  word('pizza', '🍕', 'Pizza', 'Pizza', 'بيتزا'),
  word('burger', '🍔', 'Burger', 'Burger', 'برغر'),
  word('popcorn', '🍿', 'Popcorn', 'Pop-corn', 'فشار'),
  word('cookie', '🍪', 'Cookie', 'Biscuit', 'بسكويت'),
  word('cake', '🍰', 'Cake', 'Gâteau', 'كيك'),
  word('croissant', '🥐', 'Croissant', 'Croissant', 'كرواسان'),
  word('icecream', '🍦', 'Ice cream', 'Glace', 'آيس كريم'),
  word('chocolate', '🍫', 'Chocolate', 'Chocolat', 'شوكولا'),
  word('balloon', '🎈', 'Balloon', 'Ballon', 'بالون'),
  word('gift', '🎁', 'Gift', 'Cadeau', 'هدية'),
  word('guitar', '🎸', 'Guitar', 'Guitare', 'غيتار'),
  word('umbrella', '☔', 'Umbrella', 'Parapluie', 'مظلة'),
  word('trophy', '🏆', 'Trophy', 'Trophée', 'كأس'),
  word('bicycle', '🚲', 'Bicycle', 'Vélo', 'دراجة'),
];

const BY_KEY = new Map(MATCH_WORDS.map((entry) => [entry.key, entry]));

type WordLanguage = 'en' | 'fr' | 'ar';

function wordLanguage(language: string | undefined): WordLanguage {
  const base = (language ?? '').toLowerCase().split('-')[0];
  return base === 'ar' || base === 'fr' ? base : 'en';
}

/** What a phone shows the cashier until its ticket is labelled. */
export type Match =
  | { kind: 'word'; entry: MatchWord; number: number }
  /** A ticket opened before words keeps its two character code. */
  | { kind: 'code'; code: string };

export function matchOf(ticket: {
  word?: string | undefined;
  wordNumber?: number | undefined;
  code?: string | undefined;
}): Match | null {
  if (ticket.word) {
    // A key this build does not know yet still reads as a word.
    const entry = BY_KEY.get(ticket.word) ?? {
      key: ticket.word,
      emoji: '',
      en: ticket.word,
      fr: ticket.word,
      ar: ticket.word,
    };
    return { kind: 'word', entry, number: ticket.wordNumber ?? 0 };
  }
  return ticket.code ? { kind: 'code', code: ticket.code } : null;
}

/** The word in a language, with its number when it has one ("Panda 2"). */
export function wordIn(match: Match, language: string | undefined): string {
  if (match.kind === 'code') return match.code;
  const text = match.entry[wordLanguage(language)];
  return match.number >= 2 ? `${text} ${match.number}` : text;
}

/**
 * The language a word is also shown in, smaller, so a customer and a cashier
 * who read different languages still say the same word: Arabic beside English
 * or French, English beside Arabic.
 */
export function otherWordLanguage(language: string | undefined): WordLanguage {
  return wordLanguage(language) === 'ar' ? 'en' : 'ar';
}

/** The emoji and word as one line, for titles and accessible names. */
export function matchName(match: Match, language: string | undefined): string {
  const text = wordIn(match, language);
  return match.kind === 'word' && match.entry.emoji ? `${match.entry.emoji} ${text}` : text;
}
