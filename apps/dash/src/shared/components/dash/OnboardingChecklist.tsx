import React from 'react';
import { Check, Circle, ArrowRight } from 'lucide-react';
import { motion } from 'motion/react';
import { useNavigate } from '@tanstack/react-router';
import { APP_URLS } from '@voxire/config';
import { safeRatio } from '@voxire/api';

export interface OnboardingStep {
  id: string;
  label: string;
  description: string;
  href: string;
  done: boolean;
}

export interface OnboardingChecklistProps {
  steps?: OnboardingStep[];
  onDismiss?: () => void;
}

const DEFAULT_STEPS: OnboardingStep[] = [
  {
    id: 'pos',
    label: 'Set up Point of Sale',
    description: 'Configure your first register and kitchen display.',
    href: APP_URLS.pos,
    done: false,
  },
  {
    id: 'team',
    label: 'Invite your team',
    description: 'Add your first staff member and assign apps.',
    href: '/team/invite',
    done: false,
  },
  {
    id: 'bank',
    label: 'Connect your bank',
    description: 'Link your bank account to accept payouts.',
    href: '/integrations',
    done: false,
  },
  {
    id: 'brand',
    label: 'Add your brand',
    description: 'Upload your logo and choose brand colors.',
    href: '/theme',
    done: false,
  },
  {
    id: 'menu',
    label: 'Build your menu',
    description: 'Add your first menu and a handful of items.',
    href: '/menus',
    done: false,
  },
];

export const OnboardingChecklist: React.FC<OnboardingChecklistProps> = ({
  steps = DEFAULT_STEPS,
  onDismiss,
}) => {
  const navigate = useNavigate();
  const doneCount = steps.filter((s) => s.done).length;
  const total = steps.length;
  const progress = Math.round(safeRatio(doneCount, total) * 100);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.22 }}
      className="rounded-xl p-5"
      style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
    >
      <div className="flex items-start justify-between mb-4">
        <div>
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            Get started
          </p>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            {doneCount} of {total} steps completed
          </p>
        </div>
        {onDismiss && (
          <button
            onClick={onDismiss}
            className="text-[11px] hover:opacity-70 transition-opacity"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            Dismiss
          </button>
        )}
      </div>

      <div
        className="mb-4 h-1 rounded-full overflow-hidden"
        style={{ background: 'var(--dash-hover)' }}
      >
        <motion.div
          className="h-full"
          style={{ background: '#4583C4' }}
          initial={{ width: 0 }}
          animate={{ width: `${progress}%` }}
          transition={{ duration: 0.4 }}
        />
      </div>

      <div className="space-y-1">
        {steps.map((step) => {
          const isExternal = step.href.startsWith('http');
          return (
            <button
              key={step.id}
              onClick={() => {
                if (isExternal) {
                  window.open(step.href, '_blank', 'noopener');
                } else {
                  navigate({ to: step.href });
                }
              }}
              className="group flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-left transition-colors hover:bg-[var(--dash-hover)]"
            >
              <div
                className="h-5 w-5 rounded-full flex items-center justify-center shrink-0"
                style={{
                  background: step.done ? '#10b981' : 'var(--dash-hover)',
                  border: step.done ? 'none' : '1px solid var(--dash-border)',
                }}
              >
                {step.done ? (
                  <Check size={11} color="white" strokeWidth={3} />
                ) : (
                  <Circle size={10} style={{ color: 'var(--dash-text-dim)' }} />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p
                  className="text-xs font-semibold"
                  style={{
                    color: 'var(--dash-text)',
                    textDecoration: step.done ? 'line-through' : undefined,
                    opacity: step.done ? 0.6 : 1,
                  }}
                >
                  {step.label}
                </p>
                <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {step.description}
                </p>
              </div>
              <ArrowRight
                size={12}
                className="shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
                style={{ color: 'var(--dash-text-dim)' }}
              />
            </button>
          );
        })}
      </div>
    </motion.div>
  );
};

export default OnboardingChecklist;
