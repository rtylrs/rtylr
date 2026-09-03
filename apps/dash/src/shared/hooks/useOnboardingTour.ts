import { useEffect, useRef } from 'react';
import Shepherd, { type StepOptions, type Tour } from 'shepherd.js';
import 'shepherd.js/dist/css/shepherd.css';
import { useUserStore } from '../stores/userStore';
import { onboardingClient } from '../lib/clients';

const TOUR_STEPS: StepOptions[] = [
  {
    id: 'welcome',
    text: 'Welcome to rtylr! Let me show you around your command center.',
    buttons: [
      {
        text: 'Skip tour',
        action: function (this: Tour) {
          this.complete();
        },
        secondary: true,
      },
      {
        text: 'Next',
        action: function (this: Tour) {
          this.next();
        },
      },
    ],
  },
  {
    id: 'dashboard',
    attachTo: { element: '[data-tour="dashboard"]', on: 'right' },
    text: 'This is your Dashboard. See your KPIs, recent activity, and changelog at a glance.',
    buttons: [
      {
        text: 'Back',
        action: function (this: Tour) {
          this.back();
        },
        secondary: true,
      },
      {
        text: 'Next',
        action: function (this: Tour) {
          this.next();
        },
      },
    ],
  },
  {
    id: 'team',
    attachTo: { element: '[data-tour="team"]', on: 'right' },
    text: 'Manage your team here. Add members, assign roles, and control which apps they can access.',
    buttons: [
      {
        text: 'Back',
        action: function (this: Tour) {
          this.back();
        },
        secondary: true,
      },
      {
        text: 'Next',
        action: function (this: Tour) {
          this.next();
        },
      },
    ],
  },
  {
    id: 'apps',
    attachTo: { element: '[data-tour="apps"]', on: 'right' },
    text: 'All your apps live here. POS, ERP, HR, CRM, Finance, and Flow. Click any app to launch it.',
    buttons: [
      {
        text: 'Back',
        action: function (this: Tour) {
          this.back();
        },
        secondary: true,
      },
      {
        text: 'Next',
        action: function (this: Tour) {
          this.next();
        },
      },
    ],
  },
  {
    id: 'caret',
    attachTo: { element: '[data-tour="caret"]', on: 'bottom' },
    text: 'Meet Caret, your AI assistant. Ask it anything about your business data.',
    buttons: [
      {
        text: 'Back',
        action: function (this: Tour) {
          this.back();
        },
        secondary: true,
      },
      {
        text: 'Next',
        action: function (this: Tour) {
          this.next();
        },
      },
    ],
  },
  {
    id: 'settings',
    attachTo: { element: '[data-tour="settings"]', on: 'right' },
    text: 'Account settings, billing, and integrations are all here.',
    buttons: [
      {
        text: 'Back',
        action: function (this: Tour) {
          this.back();
        },
        secondary: true,
      },
      {
        text: 'Done',
        action: function (this: Tour) {
          this.complete();
        },
      },
    ],
  },
];

export function useOnboardingTour() {
  const { user } = useUserStore();
  const tourRef = useRef<Tour | null>(null);
  const userId = user?.id;
  const hasOnboarded = user?.onboarded;

  useEffect(() => {
    if (!userId || hasOnboarded) return;
    const storageKey = `rtylr-tour-done:${userId}`;
    if (localStorage.getItem(storageKey) === '1') return;
    if (tourRef.current) return;

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

      TOUR_STEPS.forEach((step) => tour.addStep(step));

      const persistCompletion = () => {
        localStorage.setItem(storageKey, '1');
        void onboardingClient.markOnboarded({});
        tourRef.current = null;
      };

      tour.on('complete', persistCompletion);
      tour.on('cancel', persistCompletion);

      tour.start();
      tourRef.current = tour;
    }, 1000);

    return () => {
      clearTimeout(timer);
      const activeTour = tourRef.current;
      tourRef.current = null;
      activeTour?.cancel();
    };
  }, [hasOnboarded, userId]);
}
