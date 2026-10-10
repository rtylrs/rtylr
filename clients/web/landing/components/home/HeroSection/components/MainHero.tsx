import React, { useState } from 'react';
import { cn } from '../../../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { ArrowUp, Check, ChevronRight, Shuffle } from 'lucide-react';
import { CaretMark } from '../../../CaretMark/CaretMark';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { SplitWords } from '../../../motion/SplitWords';
import { SPRING_SNAPPY } from '../../../../lib/animations';
import { localeFromPath, navPath } from '../../../../lib/locales';
import { track, EV_SIGNUP_CLICK } from '../../../../lib/analytics';
import { useTypedPlaceholder } from '../../../../hooks/useTypedPlaceholder';
import ProductPreview from './ProductPreview';
import styles from '../HeroSection.module.scss';

const MainHero: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const [prompt, setPrompt] = useState('');
  const [focused, setFocused] = useState(false);

  const examples = t('home.hero.promptExamples', { returnObjects: true });
  const exampleList = Array.isArray(examples) ? (examples as string[]) : [];
  const placeholder = useTypedPlaceholder(
    exampleList,
    t('home.hero.promptPlaceholder'),
    focused || prompt.length > 0,
  );

  const fillExample = (): void => {
    if (exampleList.length === 0) return;
    const next = exampleList[Math.floor(Math.random() * exampleList.length)];
    setPrompt(next);
    document.getElementById('hero-prompt')?.focus();
  };

  const goToRegister = (): void => {
    track(EV_SIGNUP_CLICK, { source: 'hero' });
    const trimmed = prompt.trim();
    const target = trimmed
      ? `${APP_URLS.auth}/register?prompt=${encodeURIComponent(trimmed)}`
      : `${APP_URLS.auth}/register`;
    navigateExternal(target);
  };

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
    e.preventDefault();
    goToRegister();
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      goToRegister();
    }
  };

  return (
    <div className={styles.mainHero}>
      {/* Announcement chip: beta badge, one line of news, chevron */}
      <motion.a
        href={navPath('/pricing', locale)}
        className={styles.announceRow}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05, duration: 0.55, ease: [0.16, 1, 0.3, 1] }}
      >
        <span className={styles.announcePill}>{t('home.hero.announcePill')}</span>
        <span className={styles.announceText}>{t('home.hero.announceText')}</span>
        <ChevronRight className={styles.announceArrow} size={14} aria-hidden="true" />
      </motion.a>

      {/* Headline - per-word masked rise */}
      <SplitWords
        as="h1"
        className={styles.heroTitle}
        text={t('home.hero.title')}
        restLinesClassName={styles.heroTitleMuted}
        delay={0.18}
      />

      {/* Caret composer */}
      <motion.form
        className={styles.aiCard}
        onSubmit={handleSubmit}
        initial={{ opacity: 0, y: 32 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.45, duration: 0.95, ease: [0.16, 1, 0.3, 1] }}
      >
        <label htmlFor="hero-prompt" className={styles.srOnly}>
          {t('home.hero.promptLabel')}
        </label>
        <textarea
          id="hero-prompt"
          className={styles.aiCardTextarea}
          placeholder={placeholder}
          rows={3}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
        />

        <div className={styles.aiCardBottom}>
          <div className={styles.aiCardIcons}>
            <span className={cn(styles.aiCardChip, styles.aiCardChipCaret)}>
              <CaretMark size={15} />
              {t('home.hero.caretLabel')}
            </span>
            <button type="button" onClick={fillExample} className={styles.aiCardChip}>
              <Shuffle size={13} strokeWidth={2} aria-hidden="true" />
              {t('home.hero.moreOptions')}
            </button>
          </div>

          <motion.button
            type="submit"
            className={styles.aiCardSubmit}
            whileHover={{ scale: 1.03 }}
            whileTap={{ scale: 0.97 }}
            transition={SPRING_SNAPPY}
          >
            <span>{t('home.hero.submit')}</span>
            <ArrowUp size={15} strokeWidth={2.4} aria-hidden="true" />
          </motion.button>
        </div>
      </motion.form>

      {/* Beta reassurance, links straight to sign-up */}
      <motion.p
        className={styles.demoLine}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.85, duration: 0.7 }}
      >
        <Check size={14} strokeWidth={2.4} aria-hidden="true" />
        {t('home.hero.demoLinePrefix')}{' '}
        {/* APP_URLS.auth resolves to the local auth app in the browser on
            localhost and to production on the server, so the href can differ
            between the two renders in dev only. */}
        <a
          href={`${APP_URLS.auth}/register`}
          suppressHydrationWarning
          className={styles.demoLineAnchor}
          onClick={(e) => {
            e.preventDefault();
            goToRegister();
          }}
        >
          {t('home.hero.demoLineLink')}
        </a>
      </motion.p>

      <ProductPreview alt={t('home.hero.productAlt')} />
    </div>
  );
};

export default MainHero;
