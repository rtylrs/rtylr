import React, { useState } from 'react';
import { cn } from '../../../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { ArrowRight, Sparkles, Shuffle } from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { SplitWords } from '../../../motion/SplitWords';
import { SPRING_SNAPPY } from '../../../../lib/animations';
import { localeFromPath, navPath } from '../../../../lib/locales';
import styles from '../HeroSection.module.scss';

const MainHero: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const [prompt, setPrompt] = useState('');

  const goToRegister = (): void => {
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
      {/* Retool-style framing rails, scoped to just this block so they don't
          bleed into the stats section below */}
      <div className={styles.heroFrame} aria-hidden="true">
        <span className={cn(styles.heroVLine, styles.heroVLineLeft)} />
        <span className={cn(styles.heroVLine, styles.heroVLineRight)} />
      </div>

      {/* Announcement row, BETA pill sits next to plain link text */}
      <motion.a
        href={navPath('/pricing', locale)}
        className={styles.announceRow}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05, duration: 0.55, ease: [0.16, 1, 0.3, 1] }}
      >
        <span className={styles.announcePill}>{t('home.hero.announcePill')}</span>
        <span className={styles.announceText}>
          {t('home.hero.announceText')}
          <span className={styles.announceArrow} aria-hidden="true">
            ↗
          </span>
        </span>
      </motion.a>

      {/* Headline - per-word masked rise */}
      <SplitWords
        as="h1"
        className={styles.heroTitle}
        text={t('home.hero.title')}
        delay={0.18}
      />

      {/* AI prompt card, Retool AppGen clone */}
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
          placeholder={t('home.hero.promptPlaceholder')}
          rows={5}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={handleKeyDown}
        />

        <div className={styles.aiCardBottom}>
          <div className={styles.aiCardIcons}>
            <button
              type="button"
              onClick={goToRegister}
              className={cn(styles.aiCardIcon, styles.aiCardIconActive)}
              aria-label={t('home.hero.caretLabel')}
              title={t('home.hero.caretLabel')}
            >
              <Sparkles size={14} strokeWidth={2} />
            </button>
            <button
              type="button"
              onClick={goToRegister}
              className={styles.aiCardIcon}
              aria-label={t('home.hero.moreOptions')}
              title={t('home.hero.moreOptions')}
            >
              <Shuffle size={14} strokeWidth={2} />
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
            <ArrowRight size={15} strokeWidth={2.2} aria-hidden="true" />
          </motion.button>
        </div>
      </motion.form>

      {/* Talk with an expert */}
      <motion.p
        className={styles.demoLine}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.85, duration: 0.7 }}
      >
        {t('home.hero.demoLinePrefix')}{' '}
        <a href="/demo" className={styles.demoLineAnchor}>
          {t('home.hero.demoLineLink')}
        </a>
      </motion.p>
    </div>
  );
};

export default MainHero;
