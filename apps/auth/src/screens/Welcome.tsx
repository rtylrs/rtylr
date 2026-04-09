import React from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { LogIn, UserPlus, Sparkles } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import AuthDivider from '../components/AuthDivider';
import HelpLink from '../components/HelpLink';

const fieldStagger = {
  hidden: { opacity: 0, y: 12 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: 0.15 + i * 0.08, duration: 0.55, ease: [0.16, 1, 0.3, 1] },
  }),
};

const Welcome: React.FC = () => {
  const navigate = useNavigate();

  return (
    <AuthShell
      title="Welcome to rtylr."
      subtitle="The unified platform for everything your business runs on."
      footer={
        <div className="flex justify-center">
          <HelpLink />
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        <motion.div
          custom={0}
          initial="hidden"
          animate="show"
          variants={fieldStagger}
          className="rounded-xl border border-white/10 bg-white/[0.03] p-4"
        >
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-[#64C5B9]/15 text-[#64C5B9] border border-[#64C5B9]/30">
              <Sparkles className="h-5 w-5" />
            </div>
            <div>
              <p className="text-sm font-semibold text-white">
                POS, ERP, HR, CRM, Finance, Flow
              </p>
              <p className="mt-0.5 text-xs text-white/55 leading-relaxed">
                One workspace for your entire operation.
              </p>
            </div>
          </div>
        </motion.div>

        <motion.div custom={1} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthButton onClick={() => navigate('/login')}>
            <span className="inline-flex items-center justify-center gap-2">
              <LogIn className="h-[18px] w-[18px]" />
              Sign in to rtylr
            </span>
          </AuthButton>
        </motion.div>

        <motion.div custom={2} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthDivider />
        </motion.div>

        <motion.div custom={3} initial="hidden" animate="show" variants={fieldStagger}>
          <AuthButton variant="outline" onClick={() => navigate('/register')}>
            <span className="inline-flex items-center justify-center gap-2">
              <UserPlus className="h-[18px] w-[18px]" />
              Get started
            </span>
          </AuthButton>
        </motion.div>
      </div>
    </AuthShell>
  );
};

export default Welcome;
