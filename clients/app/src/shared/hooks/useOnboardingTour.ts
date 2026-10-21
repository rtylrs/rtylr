import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import Shepherd, { type StepOptions, type StepOptionsButton, type Tour } from 'shepherd.js';
import 'shepherd.js/dist/css/shepherd.css';
import { useUserStore } from '../stores/userStore';
import { onboardingClient } from '../lib/clients';

/**
 * The tour is built per run rather than held in a module constant, because its
 * copy has to come from `t` and a module constant is evaluated before i18next
 * has a language.
 */
export function buildTourSteps(t: TFunction): StepOptions[] {
  const back = (): StepOptionsButton => ({
    text: t('onboarding.tour.back'),
    action(this: Tour) {
      this.back();
    },
  });
  const next = (): StepOptionsButton => ({
    text: t('onboarding.tour.next'),
    action(this: Tour) {
      this.next();
    },
  });

  return [
    {
      id: 'welcome',
      text: t('onboarding.tour.welcome'),
      buttons: [
        {
          text: t('onboarding.tour.skip'),
          action(this: Tour) {
            this.complete();
          },
          secondary: true,
        },
        next(),
      ],
    },
    {
      id: 'dashboard',
      attachTo: { element: '[data-tour="dashboard"]', on: 'right' },
      text: t('onboarding.tour.dashboard'),
      buttons: [back(), next()],
    },
    {
      id: 'team',
      attachTo: { element: '[data-tour="team"]', on: 'right' },
      text: t('onboarding.tour.team'),
      buttons: [back(), next()],
    },
    {
      id: 'apps',
      attachTo: { element: '[data-tour="apps"]', on: 'right' },
      text: t('onboarding.tour.apps'),
      buttons: [back(), next()],
    },
    {
      id: 'caret',
      attachTo: { element: '[data-tour="caret"]', on: 'bottom' },
      text: t('onboarding.tour.caret'),
      buttons: [back(), next()],
    },
    {
      id: 'settings',
      attachTo: { element: '[data-tour="settings"]', on: 'right' },
      text: t('onboarding.tour.settings'),
      buttons: [
        back(),
        {
          text: t('onboarding.tour.done'),
          action(this: Tour) {
            this.complete();
          },
        },
      ],
    },
  ];
}

export function useOnboardingTour(enabled = true) {
  const { user } = useUserStore();
  const { t } = useTranslation();
  const tourRef = useRef<Tour | null>(null);
  // `t` changes identity when the language does. Held in a ref so the effect
  // keeps running once per user rather than restarting the tour on a language
  // switch; the steps read the language in force when the tour actually opens.
  const tRef = useRef(t);
  tRef.current = t;
  const userId = user?.id;
  const hasOnboarded = user?.onboarded;

  useEffect(() => {
    if (!enabled || !userId || hasOnboarded) return;
    const storageKey = `rtylr-tour-done:${userId}`;
    if (localStorage.getItem(storageKey) === '1') return;
    if (tourRef.current) return;

    let detachCancelListener: (() => void) | null = null;

    // Small delay to let the DOM render
    const timer = setTimeout(() => {
      const tour = new Shepherd.Tour({
        useModalOverlay: true,
        defaultStepOptions: {
          classes: 'rtylr-tour',
          scrollTo: true,
          cancelIcon: { enabled: true },
        },
      });

      buildTourSteps(tRef.current).forEach((step) => tour.addStep(step));

      const persistCompletion = () => {
        localStorage.setItem(storageKey, '1');
        void onboardingClient.markOnboarded({});
        tourRef.current = null;
      };

      tour.on('complete', persistCompletion);
      tour.on('cancel', persistCompletion);
      detachCancelListener = () => {
        tour.off('cancel', persistCompletion);
      };

      tour.start();
      tourRef.current = tour;
    }, 1000);

    return () => {
      clearTimeout(timer);
      // Teardown is not a user cancel: stop persisting completion before the
      // cancel below, so a route change does not mark onboarding as done.
      detachCancelListener?.();
      const activeTour = tourRef.current;
      tourRef.current = null;
      activeTour?.cancel();
    };
  }, [enabled, hasOnboarded, userId]);
}
