import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AnimatePresence, motion } from 'motion/react';
import { PhoneFrame, ScreenCode, ScreenReady, ScreenScan, ScreenWaiting } from './PhoneMockups';
import { findScrollRoot } from './useLandingObservers';

interface StepItem {
  title: string;
  body: string;
}

const STEP_COUNT = 4;

function renderStepScreen(index: number, active: boolean, ringKey: number): React.JSX.Element {
  switch (index) {
    case 0:
      return <ScreenScan active={active} />;
    case 1:
      return <ScreenCode />;
    case 2:
      return <ScreenWaiting ahead={2} />;
    default:
      return <ScreenReady ringKey={ringKey} />;
  }
}

export const HowItWorks = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const rawSteps = t('pager.landing.how.steps', { returnObjects: true });
  const steps = Array.isArray(rawSteps) ? (rawSteps as StepItem[]) : [];
  const rawChapters = t('pager.landing.how.chapters', { returnObjects: true });
  const chapters = Array.isArray(rawChapters) ? (rawChapters as string[]) : [];

  const [activeStep, setActiveStep] = useState(0);
  const [ringKey, setRingKey] = useState(0);
  const stepRefs = useRef<Array<HTMLLIElement | null>>([]);

  useEffect(() => {
    const nodes = stepRefs.current.filter((el): el is HTMLLIElement => el !== null);
    if (nodes.length === 0) return undefined;

    const scrollRoot = findScrollRoot(nodes[0]);
    const target: EventTarget = scrollRoot ?? window;
    let frame = 0;

    // Pick whichever step beat is closest to the vertical centre of the scroll viewport.
    const measure = (): void => {
      frame = 0;
      const rootRect = scrollRoot
        ? scrollRoot.getBoundingClientRect()
        : { top: 0, height: window.innerHeight };
      const midY = rootRect.top + rootRect.height * 0.48;

      let closest = 0;
      let bestDist = Number.POSITIVE_INFINITY;
      nodes.forEach((node, index) => {
        const rect = node.getBoundingClientRect();
        const center = rect.top + rect.height * 0.5;
        const dist = Math.abs(center - midY);
        if (dist < bestDist) {
          bestDist = dist;
          closest = index;
        }
      });

      setActiveStep((prev) => {
        if (prev !== closest && closest === STEP_COUNT - 1) setRingKey((n) => n + 1);
        return closest;
      });
    };

    const onScroll = (): void => {
      if (!frame) frame = requestAnimationFrame(measure);
    };

    measure();
    target.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onScroll, { passive: true });
    return () => {
      if (frame) cancelAnimationFrame(frame);
      target.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onScroll);
    };
  }, [steps.length]);

  const jumpStep = (index: number): void => {
    if (index === STEP_COUNT - 1 && activeStep !== index) setRingKey((n) => n + 1);
    setActiveStep(index);
    stepRefs.current[index]?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  };

  return (
    <section id="how" className="rl-section rl-story" aria-labelledby="rl-how-heading">
      <div className="rl-wrap">
        <div className="rl-story-header">
          <div className="rl-section-head">
            <h2 id="rl-how-heading" className="rl-section-title">
              {t('pager.landing.how.title')}
            </h2>
            <p className="rl-section-intro">{t('pager.landing.how.intro')}</p>
          </div>

          {/* Clickable chapter tabs with spring-gliding indicator */}
          <div className="rl-chapter-bar" role="tablist" aria-label={t('pager.landing.how.title')}>
            {chapters.map((label, idx) => {
              const active = activeStep === idx;
              return (
                <button
                  key={label}
                  type="button"
                  role="tab"
                  aria-selected={active}
                  onClick={() => jumpStep(idx)}
                  className={`rl-chapter-tab ${active ? 'rl-chapter-tab--active' : ''}`}
                >
                  {active && (
                    <motion.span
                      layoutId="rl-chapter-pill"
                      className="rl-chapter-pill"
                      transition={{ type: 'spring', stiffness: 420, damping: 34 }}
                      aria-hidden="true"
                    />
                  )}
                  <span className="rl-chapter-num" dir="ltr">
                    0{idx + 1}
                  </span>
                  <span className="rl-chapter-label">{label}</span>
                </button>
              );
            })}
          </div>
        </div>

        <div className="rl-story-grid">
          <ol className="rl-story-steps">
            {steps.map((step, index) => {
              const active = activeStep === index;
              return (
                <li
                  key={step.title}
                  ref={(el) => {
                    stepRefs.current[index] = el;
                  }}
                  data-step-index={index}
                  className="rl-story-beat"
                >
                  <div
                    className={`rl-story-card ${active ? 'rl-story-card--active' : ''}`}
                    onClick={() => jumpStep(index)}
                  >
                    <span className="rl-story-mono" dir="ltr" aria-hidden="true">
                      0{index + 1} / 0{steps.length}
                    </span>
                    <h3 className="rl-story-card-title">{step.title}</h3>
                    <p className="rl-story-card-body">{step.body}</p>

                    {/* Mobile drops pinning and renders each step's screen inline. */}
                    <div className="rl-story-mobile-phone">
                      <PhoneFrame variant="compact" ready={index === STEP_COUNT - 1}>
                        {renderStepScreen(index, true, index)}
                      </PhoneFrame>
                    </div>
                  </div>
                </li>
              );
            })}
          </ol>

          <div className="rl-story-stage" aria-hidden="true">
            <div className="rl-story-sticky">
              <div className="rl-story-stage-backdrop">
                <div className="rl-story-dots">
                  {steps.map((step, index) => (
                    <span
                      key={step.title}
                      className={`rl-story-dot ${activeStep === index ? 'rl-story-dot--active' : ''}`}
                    />
                  ))}
                </div>
                <PhoneFrame ready={activeStep === STEP_COUNT - 1}>
                  <AnimatePresence mode="wait" initial={false}>
                    <motion.div
                      key={activeStep}
                      className="rl-phone-motion-fill"
                      initial={{ opacity: 0, y: 10, scale: 0.98 }}
                      animate={{ opacity: 1, y: 0, scale: 1 }}
                      exit={{ opacity: 0, y: -10, scale: 0.98 }}
                      transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
                    >
                      {renderStepScreen(activeStep, true, ringKey)}
                    </motion.div>
                  </AnimatePresence>
                </PhoneFrame>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};
