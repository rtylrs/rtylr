import React, { useState } from 'react';
import { Globe, Check } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

export interface Language {
  code: string;
  label: string;
  native: string;
}

const LANGUAGES: Language[] = [
  { code: 'en', label: 'English', native: 'English' },
  { code: 'ar', label: 'Arabic', native: 'العربية' },
  { code: 'fr', label: 'French', native: 'Français' },
  { code: 'es', label: 'Spanish', native: 'Español' },
  { code: 'de', label: 'German', native: 'Deutsch' },
  { code: 'tr', label: 'Turkish', native: 'Türkçe' },
];

export interface LanguageSwitcherProps {
  value?: string;
  onChange?: (code: string) => void;
}

const LanguageSwitcher: React.FC<LanguageSwitcherProps> = ({
  value = 'en',
  onChange,
}) => {
  const [open, setOpen] = useState(false);
  const current = LANGUAGES.find((l) => l.code === value) || LANGUAGES[0];

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-medium text-white/70 border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] hover:text-white transition-colors"
      >
        <Globe className="h-3.5 w-3.5" />
        {current.native}
      </button>
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 6, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 6, scale: 0.96 }}
            transition={{ duration: 0.15 }}
            className="absolute right-0 mt-2 w-44 rounded-xl border border-white/10 bg-[#0a0b2e]/95 backdrop-blur-md p-1.5 shadow-xl z-50"
          >
            {LANGUAGES.map((l) => (
              <button
                key={l.code}
                type="button"
                onClick={() => {
                  onChange?.(l.code);
                  setOpen(false);
                }}
                className="w-full flex items-center justify-between gap-2 rounded-lg px-3 py-2 text-sm text-white/80 hover:bg-white/[0.06] transition-colors"
              >
                <span>{l.native}</span>
                {l.code === value && (
                  <Check className="h-3.5 w-3.5 text-[#64C5B9]" />
                )}
              </button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default LanguageSwitcher;
export { LANGUAGES };
