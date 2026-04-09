import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { PartyPopper, ArrowRight, Building2, Users, ShieldCheck } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';

const onboardingSteps = [
  {
    icon: Building2,
    title: 'Set up your business',
    subtitle: 'Name, logo, and default currency',
  },
  {
    icon: Users,
    title: 'Invite your team',
    subtitle: 'Up to 10 seats on the free plan',
  },
  {
    icon: ShieldCheck,
    title: 'Secure your account',
    subtitle: 'Enable two-factor authentication',
  },
];

const SignupComplete: React.FC = () => {
  const navigate = useNavigate();

  return (
    <AuthShell
      title="Welcome to rtylr"
      subtitle="You're all set. Let's get your workspace ready."
    >
      <div className="flex flex-col gap-5">
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ type: 'spring', stiffness: 300, damping: 20 }}
          className="flex justify-center"
        >
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-[#4583C4] to-[#7E4496] text-white">
            <PartyPopper className="h-8 w-8" />
          </div>
        </motion.div>

        <div className="flex flex-col gap-2.5">
          {onboardingSteps.map(({ icon: Icon, title, subtitle }, i) => (
            <motion.div
              key={title}
              initial={{ opacity: 0, x: -12 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.1 * i + 0.2, duration: 0.4 }}
              className="flex items-start gap-3 rounded-xl border border-white/10 bg-white/[0.03] p-4"
            >
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-white/[0.05] border border-white/10">
                <Icon className="h-4 w-4 text-white/80" />
              </div>
              <div>
                <p className="text-sm font-semibold text-white">{title}</p>
                <p className="mt-0.5 text-xs text-white/55">{subtitle}</p>
              </div>
            </motion.div>
          ))}
        </div>

        <AuthButton onClick={() => navigate('/')}>
          <span className="inline-flex items-center justify-center gap-2">
            Continue to workspace
            <ArrowRight className="h-[18px] w-[18px]" />
          </span>
        </AuthButton>
      </div>
    </AuthShell>
  );
};

export default SignupComplete;
