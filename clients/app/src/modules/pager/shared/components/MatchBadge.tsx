import React from 'react';
import { useTranslation } from 'react-i18next';
import { matchColor } from '../lib/matchCode';
import { type Match, otherWordLanguage, wordIn } from '../lib/matchWords';

interface MatchBadgeProps {
  match: Match;
  color: number | undefined;
  className: string;
  /** Classes for the smaller word in the other language, when shown. */
  otherClassName?: string;
}

/**
 * The emoji and word (or an old two character code) on its colour, as the
 * phone and the board show it. A customer and a cashier who read different
 * languages still say the same word, so the word also appears in a second
 * language underneath when `otherClassName` is given.
 */
const MatchBadge = ({
  match,
  color,
  className,
  otherClassName,
}: MatchBadgeProps): React.JSX.Element => {
  const { i18n } = useTranslation('pager');
  const language = i18n.resolvedLanguage ?? i18n.language;
  const other = otherWordLanguage(language);
  const otherWord = match.kind === 'word' ? wordIn(match, other) : '';
  const showOther =
    Boolean(otherClassName) && otherWord !== '' && otherWord !== wordIn(match, language);

  return (
    <bdi className={className} style={{ background: matchColor(color), color: '#ffffff' }}>
      {match.kind === 'word' && match.entry.emoji && (
        <span aria-hidden="true">{match.entry.emoji} </span>
      )}
      {wordIn(match, language)}
      {showOther && ' '}
      {showOther && (
        <span className={otherClassName} lang={other} dir={other === 'ar' ? 'rtl' : 'ltr'}>
          {otherWord}
        </span>
      )}
    </bdi>
  );
};

export default MatchBadge;
