import React, { useEffect, useRef, useState } from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { ArrowUp } from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { track, EV_SIGNUP_CLICK } from '../../../../lib/analytics';
import { CaretMark } from '../../../CaretMark/CaretMark';
import type { PortalView } from '../../FeaturesSection/demos/portal/PortalFrame';
import { HomeDemo } from '../../FeaturesSection/demos/HomeDemo';
import { FlowDemo } from '../../FeaturesSection/demos/FlowDemo';
import { FinanceDemo } from '../../FeaturesSection/demos/FinanceDemo';
import { CRMDemo } from '../../FeaturesSection/demos/CRMDemo';
import { HRDemo } from '../../FeaturesSection/demos/HRDemo';
import { ERPDemo } from '../../FeaturesSection/demos/ERPDemo';
import { POSDemo } from '../../FeaturesSection/demos/POSDemo';
import { InsightsDemo } from '../../FeaturesSection/demos/InsightsDemo';
import styles from './HeroPortal.module.scss';

type ViewComponent = React.ComponentType<{
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}>;

const VIEWS: Record<PortalView, ViewComponent> = {
  home: HomeDemo,
  flow: FlowDemo,
  finance: FinanceDemo,
  crm: CRMDemo,
  hr: HRDemo,
  erp: ERPDemo,
  pos: POSDemo,
  insights: InsightsDemo,
};

// Which app each suggested question opens, by index into
// home.hero.portal.caret.suggestions.
const SUGGESTION_VIEWS: PortalView[] = ['finance', 'hr', 'erp', 'insights'];
const REPLY_DELAY_MS = 900;

interface Message {
  from: 'caret' | 'user';
  text: string;
}

/**
 * The hero's live portal: a clickable rtylr workspace (mock data) beside a
 * Caret panel. Asking one of the suggested questions gets an answer and
 * opens the app it's about. Typing your own question goes to sign-up with
 * it, same as the main prompt.
 */
const HeroPortal: React.FC = () => {
  const { t } = useTranslation();
  const suggestions = t('home.hero.portal.caret.suggestions', { returnObjects: true }) as {
    q: string;
    a: string;
  }[];

  const [view, setView] = useState<PortalView>('home');
  const [messages, setMessages] = useState<Message[]>(() => [
    { from: 'caret', text: t('home.hero.portal.caret.greeting') },
  ]);
  const [asked, setAsked] = useState<number[]>([]);
  const [thinking, setThinking] = useState(false);
  const [draft, setDraft] = useState('');
  const threadRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<number | null>(null);

  useEffect(() => () => {
    if (timerRef.current !== null) window.clearTimeout(timerRef.current);
  }, []);

  useEffect(() => {
    threadRef.current?.scrollTo({ top: threadRef.current.scrollHeight, behavior: 'smooth' });
  }, [messages, thinking]);

  const ask = (i: number): void => {
    if (thinking) return;
    setAsked((a) => [...a, i]);
    setMessages((m) => [...m, { from: 'user', text: suggestions[i].q }]);
    setThinking(true);
    timerRef.current = window.setTimeout(() => {
      setMessages((m) => [...m, { from: 'caret', text: suggestions[i].a }]);
      setThinking(false);
      setView(SUGGESTION_VIEWS[i]);
    }, REPLY_DELAY_MS);
  };

  const submit = (e: React.FormEvent): void => {
    e.preventDefault();
    const text = draft.trim();
    track(EV_SIGNUP_CLICK, { source: 'hero-portal' });
    navigateExternal(
      text
        ? `${APP_URLS.auth}/register?prompt=${encodeURIComponent(text)}`
        : `${APP_URLS.auth}/register`,
    );
  };

  const View = VIEWS[view];
  const remaining = suggestions.map((s, i) => ({ ...s, i })).filter((s) => !asked.includes(s.i));

  return (
    <div className={styles.portal}>
      <div className={styles.stage}>
        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={view}
            className={styles.view}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.18 }}
          >
            <View isActive onNavigate={setView} />
          </motion.div>
        </AnimatePresence>
      </div>

      <aside className={styles.caret} aria-label={t('home.hero.portal.caret.label')}>
        <div className={styles.caretHead}>
          <span className={styles.caretMark}>
            <CaretMark size={13} />
          </span>
          Caret
        </div>

        <div className={styles.thread} ref={threadRef} aria-live="polite">
          {messages.map((m, i) => (
            <motion.div
              key={i}
              className={m.from === 'user' ? styles.msgUser : styles.msgCaret}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25 }}
            >
              {m.text}
            </motion.div>
          ))}
          {thinking && (
            <div className={styles.typing} aria-label={t('home.hero.portal.caret.thinking')}>
              <span />
              <span />
              <span />
            </div>
          )}
        </div>

        {remaining.length > 0 && (
          <div className={styles.suggestions}>
            {remaining.map((s) => (
              <button
                key={s.i}
                type="button"
                className={styles.suggestion}
                onClick={() => ask(s.i)}
                disabled={thinking}
              >
                {s.q}
              </button>
            ))}
          </div>
        )}

        <form className={styles.input} onSubmit={submit}>
          <input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder={t('home.hero.portal.caret.placeholder')}
            aria-label={t('home.hero.portal.caret.placeholder')}
          />
          <button type="submit" className={styles.send} aria-label={t('home.hero.submit')}>
            <ArrowUp size={13} strokeWidth={2.5} />
          </button>
        </form>
      </aside>
    </div>
  );
};

export default HeroPortal;
