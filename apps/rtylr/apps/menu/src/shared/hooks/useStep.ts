import { useState } from 'react';

interface UseStepReturn {
  step: number;
  isFirst: boolean;
  isLast: boolean;
  next: () => void;
  prev: () => void;
  goTo: (index: number) => void;
  reset: () => void;
}

/**
 * Manage a multi-step flow (e.g. checkout wizard).
 * Steps are 0-based internally; isFirst and isLast are derived from the count.
 */
export function useStep(totalSteps: number, initialStep = 0): UseStepReturn {
  const [step, setStep] = useState(initialStep);

  return {
    step,
    isFirst: step === 0,
    isLast: step === totalSteps - 1,
    next: () => setStep((s) => Math.min(s + 1, totalSteps - 1)),
    prev: () => setStep((s) => Math.max(s - 1, 0)),
    goTo: (index: number) => setStep(Math.max(0, Math.min(index, totalSteps - 1))),
    reset: () => setStep(initialStep),
  };
}
